# API Server - Go

Servidor Go com throttling e rate limiting configurável.

## 🚀 Tecnologias

- **Go 1.21+**
- **PostgreSQL Driver**: `github.com/lib/pq`
- **Rate Limiting**: `golang.org/x/time/rate`

## 📦 Estrutura

```
server/
├── main.go         # Código principal da API
├── go.mod          # Dependências Go
├── go.sum          # Checksums
├── Dockerfile      # Imagem Docker
└── .gitignore
```

## 🔧 Como Funciona

### Middleware de Throttling

Adiciona delay artificial antes de processar cada requisição:

```go
func throttleMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if config.ThrottleMaxMs > 0 {
            var delay int
            if config.ThrottleMinMs == config.ThrottleMaxMs {
                delay = config.ThrottleMinMs
            } else {
                // Random delay entre min e max
                delay = config.ThrottleMinMs + (int(time.Now().UnixNano()) % (config.ThrottleMaxMs - config.ThrottleMinMs + 1))
            }
            time.Sleep(time.Duration(delay) * time.Millisecond)
        }
        next(w, r)
    }
}
```

### Middleware de Rate Limiting

Usa token bucket algorithm para limitar requisições:

```go
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            w.WriteHeader(http.StatusTooManyRequests)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Rate limit exceeded",
            })
            return
        }
        next(w, r)
    }
}
```

## 🏃 Executando Localmente (Desenvolvimento)

### Com Docker (Recomendado)

```bash
# Da raiz do projeto
docker-compose up -d
```

### Sem Docker

```bash
# Instalar dependências
cd server
go mod download

# Configurar variáveis de ambiente
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=apidb
export PORT=8888
export RATE_LIMIT_REQUESTS=10
export RATE_LIMIT_PERIOD=1
export THROTTLE_MIN_MS=100
export THROTTLE_MAX_MS=500

# Executar (precisa de PostgreSQL rodando)
go run main.go
```

### Build Local

```bash
cd server
go build -o api-server main.go
./api-server
```

## 🧪 Testando

```bash
# Health check
curl http://localhost:8888/health

# Endpoint simples
curl http://localhost:8888/api/get
```

## 📊 Configuração via Variáveis de Ambiente

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `PORT` | `8888` | Porta do servidor |
| `DB_HOST` | `postgres` | Host do PostgreSQL |
| `DB_PORT` | `5432` | Porta do PostgreSQL |
| `DB_USER` | `postgres` | Usuário do banco |
| `DB_PASSWORD` | `postgres` | Senha do banco |
| `DB_NAME` | `apidb` | Nome do banco |
| `RATE_LIMIT_REQUESTS` | `10` | Número de requests permitidas |
| `RATE_LIMIT_PERIOD` | `1` | Período em segundos |
| `THROTTLE_MIN_MS` | `0` | Delay mínimo em ms |
| `THROTTLE_MAX_MS` | `0` | Delay máximo em ms |

## 🐳 Docker

### Build

```bash
cd server
docker build -t api-throttling:latest .
```

### Run

```bash
docker run -p 8888:8888 \
  -e DB_HOST=host.docker.internal \
  -e RATE_LIMIT_REQUESTS=10 \
  -e THROTTLE_MIN_MS=100 \
  -e THROTTLE_MAX_MS=500 \
  api-throttling:latest
```

## 📝 Endpoints Implementados

- `GET /health` - Health check
- `GET /api/get` - Endpoint GET simples
- `POST /api/post` - Endpoint POST com payload
- `GET /api/db/messages` - Lista mensagens do banco
- `POST /api/db/messages` - Salva mensagem no banco

## 🔄 Fluxo de Requisição

```
Request → Throttling Middleware → Rate Limit Middleware → Handler → Response
            ↓ (delay)                ↓ (check limit)        ↓ (process)
         100-500ms                  Allow/Deny              Business Logic
```

---

Voltar para o [README principal](../README.md)

