# API Throttling - Simulador de Rate Limiting e Throttling

API REST simples em Go para simular **throttling** (latência artificial) e **rate limiting** (limitação de requisições) com suporte a banco de dados PostgreSQL.

## 🚀 Características

- ✅ **Rate Limiting Configurável**: Controle de requisições por segundo (ex: máximo 10 req/s)
- ✅ **Throttling/Latência Artificial**: Adiciona delay configurável (ex: 100-500ms por request)
- ✅ **Endpoints Simples**: GET e POST para testes básicos
- ✅ **Integração com PostgreSQL**: Endpoints para consulta e gravação no banco
- ✅ **Alto Desempenho**: Implementado em Go para máxima performance
- ✅ **Docker Compose**: Ambiente completo containerizado
- ✅ **Clientes de Teste**: Scripts em Bash e Python prontos para uso

## 📖 Diferença entre Rate Limiting e Throttling

### Rate Limiting
Limita o **número de requisições** em um período. Quando o limite é atingido, requisições adicionais são **rejeitadas** com erro 429.

**Exemplo**: 10 requisições por segundo
- Requisições 1-10: ✅ Processadas
- Requisição 11: ❌ Rejeitada (429 Too Many Requests)

### Throttling
Adiciona **delay/latência artificial** a cada requisição para simular servidores lentos ou redes instáveis. Todas as requisições são processadas, mas com atraso.

**Exemplo**: 100-500ms de delay por requisição
- Cada requisição aguarda um tempo aleatório entre 100ms e 500ms antes de ser processada

---

## 📋 Quick Start (2 minutos)

### 1️⃣ Subir a API

```bash
docker-compose up -d
```

Aguarde ~10 segundos para tudo inicializar.

### 2️⃣ Testar

#### Opção A: Script Automatizado (Bash)
```bash
cd client-tests
./test-api.sh
```

#### Opção B: Cliente Python
```bash
cd client-tests
pip3 install -r requirements.txt
python3 example-client.py
```

#### Opção C: Manualmente com curl
```bash
# Health check (ver status completo)
curl http://localhost:8888/health | jq '.'

# Ver configuração ativa
curl -s http://localhost:8888/health | jq '.configuration'

# GET simples
curl http://localhost:8888/api/get

# POST com payload
curl -X POST http://localhost:8888/api/post \
  -H "Content-Type: application/json" \
  -d '{"test":"hello","value":123}'

# Salvar no banco
curl -X POST http://localhost:8888/api/db/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"Minha primeira mensagem"}'

# Listar do banco
curl http://localhost:8888/api/db/messages
```

### 3️⃣ Ver Rate Limiting em Ação

```bash
# 20 requisições rápidas - algumas serão rejeitadas (HTTP 429)
for i in {1..20}; do
  curl -s -w "HTTP %{http_code}\n" http://localhost:8888/api/get
done
```

### 4️⃣ Ver Throttling em Ação

```bash
# Medir tempo - deve levar ~100-500ms por request
time curl http://localhost:8888/api/get
```

---

## 🔧 Configuração

### Variáveis de Ambiente

Configure o rate limiting e throttling editando as variáveis no `docker-compose.yml`:

```yaml
environment:
  # Rate Limiting (limite de requisições)
  - RATE_LIMIT_REQUESTS=10  # Número de requisições permitidas
  - RATE_LIMIT_PERIOD=1     # Período em segundos
  
  # Throttling (latência artificial)
  - THROTTLE_MIN_MS=100     # Delay mínimo em milissegundos
  - THROTTLE_MAX_MS=500     # Delay máximo em milissegundos
```

### Exemplos de Configuração

#### Rate Limiting:
- `RATE_LIMIT_REQUESTS=10` e `RATE_LIMIT_PERIOD=1` → 10 requisições por segundo
- `RATE_LIMIT_REQUESTS=100` e `RATE_LIMIT_PERIOD=60` → 100 requisições por minuto (~1.67 req/s)
- `RATE_LIMIT_REQUESTS=1000` e `RATE_LIMIT_PERIOD=3600` → 1000 requisições por hora

#### Throttling:
- `THROTTLE_MIN_MS=100` e `THROTTLE_MAX_MS=500` → Delay aleatório entre 100-500ms
- `THROTTLE_MIN_MS=200` e `THROTTLE_MAX_MS=200` → Delay fixo de 200ms
- `THROTTLE_MIN_MS=0` e `THROTTLE_MAX_MS=0` → Throttling desabilitado (sem delay)
- `THROTTLE_MIN_MS=1000` e `THROTTLE_MAX_MS=3000` → Simula servidor muito lento (1-3 segundos)

### Cenários de Teste Pré-configurados

#### Cenário 1: API Rápida (sem limitações)
```yaml
- RATE_LIMIT_REQUESTS=1000
- RATE_LIMIT_PERIOD=1
- THROTTLE_MIN_MS=0
- THROTTLE_MAX_MS=0
```

#### Cenário 2: API Normal (pequeno delay)
```yaml
- RATE_LIMIT_REQUESTS=50
- RATE_LIMIT_PERIOD=1
- THROTTLE_MIN_MS=50
- THROTTLE_MAX_MS=150
```

#### Cenário 3: API Lenta (servidor sobrecarregado)
```yaml
- RATE_LIMIT_REQUESTS=10
- RATE_LIMIT_PERIOD=1
- THROTTLE_MIN_MS=500
- THROTTLE_MAX_MS=2000
```

#### Cenário 4: API Muito Restritiva
```yaml
- RATE_LIMIT_REQUESTS=5
- RATE_LIMIT_PERIOD=10
- THROTTLE_MIN_MS=1000
- THROTTLE_MAX_MS=3000
```

Após alterar, reinicie:
```bash
docker-compose restart api
```

---

## 🐳 Comandos Docker Compose

### Usando Makefile (Recomendado)

```bash
# Ver comandos disponíveis
make help

# Iniciar tudo
make up

# Ver logs
make logs

# Executar testes
make test

# Parar tudo
make down
```

### Cenários de Teste Rápidos com Makefile

```bash
make scenario-fast      # API rápida (sem limitações)
make scenario-normal    # API normal (pequeno delay)
make scenario-slow      # API lenta (servidor sobrecarregado)
make scenario-strict    # API muito restritiva
```

### Comandos Docker Compose Diretos

```bash
# Iniciar os serviços
docker-compose up -d

# Parar os serviços
docker-compose down

# Reconstruir após mudanças
docker-compose up -d --build

# Ver logs
docker-compose logs -f api

# Ver status
docker-compose ps
```

---

## 📡 Endpoints da API

Todos os endpoints rodam em `http://localhost:8888`

> 📋 **Documentação OpenAPI**: Veja o arquivo [`openapi.yaml`](openapi.yaml) para a especificação completa da API em formato OpenAPI 3.0

### Health Check
```bash
GET /health
```

**Resposta (quando tudo está OK):**
```json
{
  "status": "ok",
  "time": "2025-11-15T10:30:00Z",
  "database": {
    "status": "connected",
    "host": "postgres",
    "port": "5432",
    "name": "apidb"
  },
  "configuration": {
    "rate_limiting": {
      "requests": 5,
      "period_seconds": 1,
      "rate_per_second": 5
    },
    "throttling": {
      "min_ms": 1000,
      "max_ms": 3000,
      "enabled": true
    }
  },
  "server": {
    "port": "8888"
  }
}
```

**Resposta (quando banco está desconectado - HTTP 503):**
```json
{
  "status": "degraded",
  "time": "2025-11-15T10:30:00Z",
  "database": {
    "status": "disconnected",
    "host": "postgres",
    "port": "5432",
    "name": "apidb",
    "error": "connection refused"
  },
  "configuration": {
    "rate_limiting": {
      "requests": 5,
      "period_seconds": 1,
      "rate_per_second": 5
    },
    "throttling": {
      "min_ms": 1000,
      "max_ms": 3000,
      "enabled": true
    }
  },
  "server": {
    "port": "8888"
  }
}
```

### GET Simples
```bash
GET /api/get
```

**Resposta:**
```json
{
  "message": "GET request received successfully",
  "time": "2025-11-15T10:30:00Z"
}
```

### POST Simples
```bash
POST /api/post
Content-Type: application/json

{
  "name": "Test",
  "value": 123,
  "data": {"nested": "object"}
}
```

**Resposta:**
```json
{
  "message": "POST request received successfully",
  "received": {
    "name": "Test",
    "value": 123,
    "data": {"nested": "object"}
  },
  "time": "2025-11-15T10:30:00Z"
}
```

### GET do Banco (Listar Mensagens)
```bash
GET /api/db/messages
```

**Resposta:**
```json
{
  "count": 2,
  "messages": [
    {
      "id": 2,
      "content": "Segunda mensagem",
      "created_at": "2025-11-15T10:31:00Z"
    },
    {
      "id": 1,
      "content": "Primeira mensagem",
      "created_at": "2025-11-15T10:30:00Z"
    }
  ]
}
```

### POST no Banco (Salvar Mensagem)
```bash
POST /api/db/messages
Content-Type: application/json

{
  "content": "Minha mensagem para salvar no banco"
}
```

**Resposta (HTTP 201):**
```json
{
  "message": "Message saved successfully",
  "data": {
    "id": 1,
    "content": "Minha mensagem para salvar no banco",
    "created_at": "2025-11-15T10:30:00Z"
  }
}
```

### Rate Limit Excedido

**Resposta (HTTP 429):**
```json
{
  "error": "Rate limit exceeded. Too many requests."
}
```

---

## 🧪 Testando com Ferramentas de Benchmark

### Apache Bench

```bash
# 100 requisições, 10 concorrentes
ab -n 100 -c 10 http://localhost:8888/api/get
```

### hey (ferramenta Go)

```bash
# Instalar hey
go install github.com/rakyll/hey@latest

# 200 requisições, 50 concorrentes
hey -n 200 -c 50 http://localhost:8888/api/get
```

### wrk

```bash
# 10 segundos, 2 threads, 10 conexões
wrk -t2 -c10 -d10s http://localhost:8888/api/get
```

---

## 🗄️ Acessando o Banco de Dados

```bash
# Conectar ao PostgreSQL
docker-compose exec postgres psql -U postgres -d apidb

# Ver mensagens
SELECT * FROM messages;

# Contar mensagens
SELECT COUNT(*) FROM messages;

# Deletar todas as mensagens
DELETE FROM messages;

# Sair
\q
```

Ou usando make:
```bash
make db
```

---

## 📊 Monitoramento

### Ver logs em tempo real:
```bash
docker-compose logs -f api
```

### Ver configuração ativa:
```bash
docker-compose logs api | grep -E "(Rate limiter|Throttling)"
```

### Ver estatísticas dos containers:
```bash
docker stats
```

### Health check:
```bash
make health
# ou
curl http://localhost:8888/health | jq '.'

# Ver apenas o status
curl -s http://localhost:8888/health | jq -r '.status'

# Ver apenas a configuração
curl -s http://localhost:8888/health | jq '.configuration'

# Ver status do banco
curl -s http://localhost:8888/health | jq '.database'
```

**O que o `/health` retorna:**
- ✅ **Status geral**: `ok` (HTTP 200) quando tudo está funcionando
- ⚠️ **Status degradado**: `degraded` (HTTP 503) quando o banco está desconectado
- 📊 **Conexão do banco**: Verifica com `db.Ping()` e mostra host/port/nome
- ⚙️ **Configurações ativas**: Valores atuais de rate limiting e throttling
- 🕐 **Timestamp**: Horário da verificação
- 🔢 **Taxa calculada**: Requisições por segundo (rate_per_second)

**Ideal para:**
- Health checks de Kubernetes/Docker
- Monitoramento (Prometheus, Datadog, etc.)
- Load balancer checks
- Verificação em CI/CD pipelines

---

## 💻 Exemplos de Código em Diferentes Linguagens

### Python

```python
import requests

# GET simples
response = requests.get("http://localhost:8888/api/get")
print(response.json())

# POST com payload
data = {"name": "Test", "value": 123}
response = requests.post(
    "http://localhost:8888/api/post",
    json=data
)
print(response.json())

# Tratando Rate Limiting com retry
def make_request_with_retry(url, max_retries=3):
    for attempt in range(max_retries):
        response = requests.get(url)
        if response.status_code == 200:
            return response.json()
        elif response.status_code == 429:
            wait_time = 2 ** attempt  # Exponential backoff
            print(f"Rate limited. Aguardando {wait_time}s...")
            time.sleep(wait_time)
        else:
            raise Exception(f"HTTP {response.status_code}")
    raise Exception("Max retries exceeded")
```

### JavaScript / Node.js

```javascript
// Com fetch (Node.js 18+)
const response = await fetch("http://localhost:8888/api/get");
const data = await response.json();
console.log(data);

// POST
const postResponse = await fetch("http://localhost:8888/api/post", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ name: "Test", value: 123 }),
});
console.log(await postResponse.json());
```

### Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func main() {
    // GET simples
    resp, _ := http.Get("http://localhost:8888/api/get")
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
    
    // POST
    payload := map[string]interface{}{"name": "Test", "value": 123}
    jsonData, _ := json.Marshal(payload)
    http.Post(
        "http://localhost:8888/api/post",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
}
```

### Bash / cURL

```bash
# GET simples
curl http://localhost:8888/api/get

# GET com formatação JSON (requer jq)
curl -s http://localhost:8888/api/get | jq '.'

# POST com payload
curl -X POST http://localhost:8888/api/post \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","value":123}'

# Medir tempo de resposta
curl -w "\nTempo: %{time_total}s\n" http://localhost:8888/api/get
```

---

## 📝 Estrutura do Projeto

```
api-throttling/
├── server/                    # Código do servidor Go
│   ├── main.go               # Código principal da API
│   ├── go.mod                # Dependências Go
│   ├── go.sum                # Checksums
│   ├── Dockerfile            # Imagem Docker
│   └── README.md             # Doc do servidor
├── client-tests/             # Clientes de teste
│   ├── test-api.sh          # Script de teste (Bash)
│   ├── example-client.py    # Cliente completo (Python)
│   ├── requirements.txt     # Dependências Python
│   └── README.md            # Doc dos testes
├── docker-compose.yml        # Orquestração dos serviços
├── Makefile                  # Comandos facilitadores
├── openapi.yaml             # 📋 Especificação OpenAPI 3.0
├── .env.example             # Exemplo de variáveis de ambiente
├── .gitignore              
└── README.md                # Esta documentação
```

---

## 🎯 Use Cases

1. **Testar Rate Limiting**: Simule diferentes cargas de requisições e veja rejeições
2. **Simular APIs Lentas**: Teste como sua aplicação se comporta com latência alta
3. **Testes de Timeout**: Verifique se seus timeouts estão configurados corretamente
4. **Testes de Retry**: Valide lógica de retry em clientes HTTP
5. **Benchmark de Performance**: Compare diferentes estratégias de paralelização
6. **Desenvolvimento Local**: Simule comportamento de APIs de produção
7. **QA/Testes de Carga**: Valide comportamento sob diferentes condições de rede
8. **Demonstrações**: Mostre diferença entre sistemas rápidos e lentos
9. **Treinamento**: Aprenda sobre rate limiting, throttling e Go

---

## 🛠️ Desenvolvimento Local (sem Docker)

```bash
# Iniciar apenas o PostgreSQL com Docker
docker run -d \
  --name postgres-dev \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=apidb \
  -p 5432:5432 \
  postgres:16-alpine

# Configurar variáveis de ambiente
export DB_HOST=localhost
export PORT=8888
export RATE_LIMIT_REQUESTS=10
export RATE_LIMIT_PERIOD=1
export THROTTLE_MIN_MS=100
export THROTTLE_MAX_MS=500

# Executar a aplicação
cd server
go run main.go
```

---

## ❓ Troubleshooting

### API não responde
```bash
docker-compose ps        # Ver status dos containers
docker-compose logs api  # Ver logs da API
```

### Porta 8888 ocupada
Edite `docker-compose.yml` e mude:
```yaml
ports:
  - "8888:8888"  # Mude para "9999:8888" por exemplo
```

### Resetar tudo
```bash
docker-compose down -v    # Remove containers e volumes
docker-compose up -d      # Sobe novamente
```

### Ver configuração atual
```bash
docker-compose logs api | grep -E "(Rate limiter|Throttling)"
```

---

## 🔐 Segurança

⚠️ **Esta é uma aplicação de demonstração**. Para uso em produção:

- Use senhas fortes e seguras
- Configure SSL/TLS (HTTPS)
- Adicione autenticação e autorização
- Use secrets management (não variáveis de ambiente em texto claro)
- Configure rate limiting por IP ou usuário
- Adicione logging e monitoring adequados
- Implemente CORS apropriadamente
- Use helmet ou equivalente para headers de segurança

---

## 📦 Tecnologias Utilizadas

- **Go 1.21+**: Linguagem principal
- **PostgreSQL 16**: Banco de dados
- **Docker & Docker Compose**: Containerização
- **golang.org/x/time/rate**: Rate limiting
- **lib/pq**: Driver PostgreSQL para Go

---

## 📄 Licença

MIT License - Sinta-se livre para usar e modificar.

---

## 🤝 Contribuindo

Sugestões e melhorias são bem-vindas! Este é um projeto educacional focado em demonstrar conceitos de rate limiting e throttling.

---

**Desenvolvido com ❤️ em Go**
