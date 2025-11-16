package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/time/rate"
)

var (
	db      *sql.DB
	limiter *rate.Limiter
	config  Config
)

type Config struct {
	Port              string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	RateLimitRequests int
	RateLimitPeriod   int // seconds
	ThrottleMinMs     int // minimum delay in milliseconds
	ThrottleMaxMs     int // maximum delay in milliseconds
}

type Message struct {
	ID        int       `json:"id,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func loadConfig() Config {
	rateLimitRequests, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "10"))
	rateLimitPeriod, _ := strconv.Atoi(getEnv("RATE_LIMIT_PERIOD", "1"))
	throttleMinMs, _ := strconv.Atoi(getEnv("THROTTLE_MIN_MS", "0"))
	throttleMaxMs, _ := strconv.Atoi(getEnv("THROTTLE_MAX_MS", "0"))

	return Config{
		Port:              getEnv("PORT", "8888"),
		DBHost:            getEnv("DB_HOST", "postgres"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "apidb"),
		RateLimitRequests: rateLimitRequests,
		RateLimitPeriod:   rateLimitPeriod,
		ThrottleMinMs:     throttleMinMs,
		ThrottleMaxMs:     throttleMaxMs,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initDB(config Config) error {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	log.Printf("[DB] Connecting to PostgreSQL at %s:%s...", config.DBHost, config.DBPort)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("[DB] Error opening connection: %v", err)
		return err
	}

	// Configurar pool de conexões para ALTA performance (10k+ TPS)
	db.SetMaxOpenConns(200) // 200 conexões simultâneas
	db.SetMaxIdleConns(200) // Manter todas idle ativas
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)
	log.Printf("[DB] Connection pool configured: MaxOpen=200, MaxIdle=200, MaxLifetime=5m, IdleTime=1m")

	// Wait for database to be ready
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			log.Printf("[DB] Connection successful!")
			break
		}
		log.Printf("[DB] Waiting for database... (%d/%d) - Error: %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Printf("[DB] Failed to connect after %d retries: %v", maxRetries, err)
		return err
	}

	// Create table if not exists
	log.Printf("[DB] Creating tables if not exist...")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)

	if err != nil {
		log.Printf("[DB] Error creating tables: %v", err)
		return err
	}

	log.Printf("[DB] Tables ready")
	return nil
}

func throttleMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apply artificial delay (throttling)
		if config.ThrottleMaxMs > 0 {
			var delay int
			if config.ThrottleMinMs == config.ThrottleMaxMs {
				delay = config.ThrottleMinMs
			} else {
				// Random delay between min and max
				delay = config.ThrottleMinMs + (int(time.Now().UnixNano()) % (config.ThrottleMaxMs - config.ThrottleMinMs + 1))
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		next(w, r)
	}
}

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Rate limit exceeded. Too many requests.",
			})
			return
		}
		next(w, r)
	}
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// OTIMIZAÇÃO: Logs desabilitados para alta performance
		// Descomentar apenas para debug (impacta TPS significativamente)

		// start := time.Now()
		// log.Printf("[REQUEST] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		next(w, r)

		// duration := time.Since(start)
		// log.Printf("[RESPONSE] %s %s completed in %v", r.Method, r.URL.Path, duration)
	}
}

func combinedMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return loggingMiddleware(throttleMiddleware(rateLimitMiddleware(next)))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	log.Printf("[HEALTH] Health check request from %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")

	// Verificar conexão com o banco
	dbStatus := "connected"
	dbError := ""
	pingStart := time.Now()
	if err := db.Ping(); err != nil {
		dbStatus = "disconnected"
		dbError = err.Error()
		log.Printf("[HEALTH] Database ping failed in %v: %v", time.Since(pingStart), err)
	} else {
		log.Printf("[HEALTH] Database ping successful in %v", time.Since(pingStart))
	}

	response := map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
		"database": map[string]interface{}{
			"status": dbStatus,
			"host":   config.DBHost,
			"port":   config.DBPort,
			"name":   config.DBName,
		},
		"configuration": map[string]interface{}{
			"rate_limiting": map[string]interface{}{
				"requests":        config.RateLimitRequests,
				"period_seconds":  config.RateLimitPeriod,
				"rate_per_second": float64(config.RateLimitRequests) / float64(config.RateLimitPeriod),
			},
			"throttling": map[string]interface{}{
				"min_ms":  config.ThrottleMinMs,
				"max_ms":  config.ThrottleMaxMs,
				"enabled": config.ThrottleMaxMs > 0,
			},
		},
		"server": map[string]interface{}{
			"port": config.Port,
		},
	}

	// Se houver erro no banco, adicionar detalhes
	if dbError != "" {
		response["database"].(map[string]interface{})["error"] = dbError
		response["status"] = "degraded"
	}

	// Status code baseado na saúde
	if dbStatus == "disconnected" {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Printf("[HEALTH] Returning 503 (degraded) - DB disconnected")
	}

	json.NewEncoder(w).Encode(response)
	log.Printf("[HEALTH] Health check completed in %v", time.Since(start))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "GET request received successfully",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON payload",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "POST request received successfully",
		"received": payload,
		"time":     time.Now().Format(time.RFC3339),
	})
}

func dbGetHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, content, created_at FROM messages ORDER BY created_at DESC LIMIT 100")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Database query failed",
		})
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Content, &msg.CreatedAt); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":    len(messages),
		"messages": messages,
	})
}

func dbPostHandler(w http.ResponseWriter, r *http.Request) {
	var msg Message

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON payload. Expected: {\"content\": \"your message\"}",
		})
		return
	}

	if msg.Content == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Content field is required",
		})
		return
	}

	var id int
	var createdAt time.Time
	err := db.QueryRow(
		"INSERT INTO messages (content) VALUES ($1) RETURNING id, created_at",
		msg.Content,
	).Scan(&id, &createdAt)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to insert message",
		})
		return
	}

	msg.ID = id
	msg.CreatedAt = createdAt

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Message saved successfully",
		"data":    msg,
	})
}

func main() {
	log.Println("==========================================")
	log.Println("  API Throttling Server Starting...")
	log.Println("  🚀 HIGH PERFORMANCE MODE - 10k+ TPS")
	log.Println("==========================================")

	// Configurar GOMAXPROCS para usar todos os CPUs disponíveis
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	log.Printf("[CONFIG] GOMAXPROCS set to %d CPUs", numCPU)

	config = loadConfig()

	// Log da configuração
	log.Printf("[CONFIG] Port: %s", config.Port)
	log.Printf("[CONFIG] Database: %s:%s/%s", config.DBHost, config.DBPort, config.DBName)

	// Initialize rate limiter
	// Rate: requests per second = RateLimitRequests / RateLimitPeriod
	ratePerSecond := float64(config.RateLimitRequests) / float64(config.RateLimitPeriod)
	limiter = rate.NewLimiter(rate.Limit(ratePerSecond), config.RateLimitRequests)

	log.Printf("[CONFIG] Rate limiter: %d requests per %d second(s) (%.2f req/s)",
		config.RateLimitRequests, config.RateLimitPeriod, ratePerSecond)

	if config.ThrottleMaxMs > 0 {
		log.Printf("[CONFIG] Throttling enabled: %d-%d ms delay per request",
			config.ThrottleMinMs, config.ThrottleMaxMs)
	} else {
		log.Printf("[CONFIG] Throttling disabled (THROTTLE_MAX_MS = 0)")
	}

	// Initialize database
	log.Println("[INIT] Initializing database connection...")
	if err := initDB(config); err != nil {
		log.Fatalf("[FATAL] Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("[INIT] Database connected successfully!")

	// Routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/get", combinedMiddleware(getHandler))
	http.HandleFunc("/api/post", combinedMiddleware(postHandler))
	http.HandleFunc("/api/db/messages", func(w http.ResponseWriter, r *http.Request) {
		combinedMiddleware(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				dbGetHandler(w, r)
			} else if r.Method == http.MethodPost {
				dbPostHandler(w, r)
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})(w, r)
	})

	log.Println("==========================================")
	log.Printf("[SERVER] Starting on port %s", config.Port)
	log.Println("[SERVER] Endpoints:")
	log.Println("  - GET  /health")
	log.Println("  - GET  /api/get")
	log.Println("  - POST /api/post")
	log.Println("  - GET  /api/db/messages")
	log.Println("  - POST /api/db/messages")
	log.Println("==========================================")
	log.Printf("[SERVER] 🚀 High Performance Server ready at http://0.0.0.0:%s", config.Port)
	log.Printf("[SERVER] 📊 Target: 10k+ TPS | %d CPUs | Pool: 200 connections", numCPU)
	log.Println("==========================================")

	// Configurar servidor HTTP para alta performance
	server := &http.Server{
		Addr:           ":" + config.Port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("[FATAL] Server failed to start: %v", err)
	}
}
