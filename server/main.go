package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	// Wait for database to be ready
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for database... (%d/%d)", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return err
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)

	return err
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

func combinedMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return throttleMiddleware(rateLimitMiddleware(next))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Verificar conexão com o banco
	dbStatus := "connected"
	dbError := ""
	if err := db.Ping(); err != nil {
		dbStatus = "disconnected"
		dbError = err.Error()
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
				"requests":      config.RateLimitRequests,
				"period_seconds": config.RateLimitPeriod,
				"rate_per_second": float64(config.RateLimitRequests) / float64(config.RateLimitPeriod),
			},
			"throttling": map[string]interface{}{
				"min_ms":    config.ThrottleMinMs,
				"max_ms":    config.ThrottleMaxMs,
				"enabled":   config.ThrottleMaxMs > 0,
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
	}
	
	json.NewEncoder(w).Encode(response)
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
	config = loadConfig()

	// Initialize rate limiter
	// Rate: requests per second = RateLimitRequests / RateLimitPeriod
	ratePerSecond := float64(config.RateLimitRequests) / float64(config.RateLimitPeriod)
	limiter = rate.NewLimiter(rate.Limit(ratePerSecond), config.RateLimitRequests)

	log.Printf("Rate limiter configured: %d requests per %d second(s)", 
		config.RateLimitRequests, config.RateLimitPeriod)
	
	if config.ThrottleMaxMs > 0 {
		log.Printf("Throttling configured: %d-%d ms delay per request", 
			config.ThrottleMinMs, config.ThrottleMaxMs)
	} else {
		log.Println("Throttling disabled (THROTTLE_MAX_MS = 0)")
	}

	// Initialize database
	if err := initDB(config); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

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

	log.Printf("Server starting on port %s", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

