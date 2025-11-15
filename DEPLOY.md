# 🚀 Guia de Deploy

Este guia mostra como fazer deploy da API Throttling em diferentes plataformas usando Nixpacks.

## 📦 Plataformas Suportadas

- ✅ **Dokploy** (self-hosted)
- ✅ **Railway**
- ✅ **Render**
- ✅ **Fly.io**
- ✅ **Heroku** (com buildpack)
- ✅ Qualquer plataforma que suporte Nixpacks ou Docker

---

## 🐳 Dokploy (Self-Hosted)

### 1. Pré-requisitos
- Dokploy instalado e rodando
- Acesso ao painel de administração
- PostgreSQL disponível (pode criar no Dokploy)

### 2. Criar Banco de Dados PostgreSQL

1. No Dokploy, vá em **Databases** → **Create Database**
2. Escolha **PostgreSQL 16**
3. Configure:
   - Name: `api-throttling-db`
   - User: `postgres`
   - Password: `[senha-segura]`
   - Database: `apidb`
4. Anote o **Internal Host** (ex: `api-throttling-db:5432`)

### 3. Deploy da Aplicação

1. **Criar Nova Aplicação**:
   - Go to **Applications** → **Create Application**
   - Name: `api-throttling`
   - Type: **Nixpacks**

2. **Configurar Repository**:
   - Repository URL: `https://github.com/seu-usuario/api-trotling`
   - Branch: `main`
   - Root Directory: `/` (deixar vazio se na raiz)

3. **Configurar Variáveis de Ambiente**:
   ```env
   PORT=8888
   DB_HOST=api-throttling-db
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=sua-senha-aqui
   DB_NAME=apidb
   RATE_LIMIT_REQUESTS=10
   RATE_LIMIT_PERIOD=1
   THROTTLE_MIN_MS=100
   THROTTLE_MAX_MS=500
   ```

4. **Configurar Porta**:
   - Port: `8888`
   - Protocol: `HTTP`

5. **Deploy**:
   - Click **Deploy**
   - Aguarde o build completar

6. **Acessar**:
   - Dokploy irá fornecer uma URL (ex: `https://api-throttling.seu-dominio.com`)
   - Teste: `curl https://api-throttling.seu-dominio.com/health`

---

## 🚂 Railway

### 1. Deploy via Dashboard

1. Acesse [Railway.app](https://railway.app)
2. **New Project** → **Deploy from GitHub**
3. Selecione seu repositório
4. Railway detectará automaticamente o `nixpacks.toml`

### 2. Adicionar PostgreSQL

1. No projeto, clique **New** → **Database** → **PostgreSQL**
2. Railway criará automaticamente as variáveis:
   - `DATABASE_URL`
   - `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`

### 3. Configurar Variáveis de Ambiente

No service da API, adicione:

```env
PORT=8888
DB_HOST=${{Postgres.PGHOST}}
DB_PORT=${{Postgres.PGPORT}}
DB_USER=${{Postgres.PGUSER}}
DB_PASSWORD=${{Postgres.PGPASSWORD}}
DB_NAME=${{Postgres.PGDATABASE}}
RATE_LIMIT_REQUESTS=10
RATE_LIMIT_PERIOD=1
THROTTLE_MIN_MS=100
THROTTLE_MAX_MS=500
```

### 4. Expor Porta

1. No service, vá em **Settings**
2. **Networking** → **Public Domain**
3. Railway gerará uma URL pública

### 5. Deploy

- Push para o repositório ou clique **Deploy**
- URL exemplo: `https://api-throttling-production.up.railway.app`

---

## 🎨 Render

### 1. Criar Web Service

1. Acesse [Render.com](https://render.com)
2. **New** → **Web Service**
3. Conecte seu repositório GitHub

### 2. Configurar Build

- **Name**: `api-throttling`
- **Environment**: `Go`
- **Build Command**: `cd server && go build -o /opt/render/project/api-server main.go`
- **Start Command**: `./api-server`

### 3. Criar PostgreSQL Database

1. **New** → **PostgreSQL**
2. **Name**: `api-throttling-db`
3. Plan: Free ou Starter
4. Anote as credenciais

### 4. Variáveis de Ambiente

No Web Service:

```env
PORT=8888
DB_HOST=dpg-xxx.oregon-postgres.render.com
DB_PORT=5432
DB_USER=api_throttling_user
DB_PASSWORD=senha-gerada-pelo-render
DB_NAME=api_throttling
RATE_LIMIT_REQUESTS=10
RATE_LIMIT_PERIOD=1
THROTTLE_MIN_MS=100
THROTTLE_MAX_MS=500
```

### 5. Deploy

- Render fará deploy automático
- URL: `https://api-throttling.onrender.com`

---

## 🪰 Fly.io

### 1. Instalar Flyctl

```bash
# macOS
brew install flyctl

# Linux
curl -L https://fly.io/install.sh | sh

# Login
flyctl auth login
```

### 2. Criar Aplicação

```bash
cd /Users/renatoruis/work/DATASTREAM/MERCANTIL/api-trotling

# Inicializar
flyctl launch --name api-throttling --region gru --no-deploy
```

### 3. Criar PostgreSQL

```bash
# Criar banco
flyctl postgres create --name api-throttling-db --region gru

# Conectar ao app
flyctl postgres attach --app api-throttling api-throttling-db
```

Fly.io criará automaticamente `DATABASE_URL`.

### 4. Configurar Secrets (Variáveis)

```bash
flyctl secrets set \
  PORT=8888 \
  RATE_LIMIT_REQUESTS=10 \
  RATE_LIMIT_PERIOD=1 \
  THROTTLE_MIN_MS=100 \
  THROTTLE_MAX_MS=500 \
  --app api-throttling
```

### 5. Criar fly.toml

```toml
app = "api-throttling"
primary_region = "gru"

[build]
  builder = "paketobuildpacks/builder:base"

[env]
  PORT = "8888"

[http_service]
  internal_port = 8888
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 0

[[services]]
  protocol = "tcp"
  internal_port = 8888

  [[services.ports]]
    port = 80
    handlers = ["http"]

  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]
```

### 6. Deploy

```bash
flyctl deploy
```

URL: `https://api-throttling.fly.dev`

---

## 🔧 Variáveis de Ambiente Requeridas

### Obrigatórias

```env
PORT=8888                    # Porta do servidor
DB_HOST=postgres-host        # Host do PostgreSQL
DB_PORT=5432                 # Porta do PostgreSQL
DB_USER=user                 # Usuário do banco
DB_PASSWORD=password         # Senha do banco
DB_NAME=apidb                # Nome do banco
```

### Opcionais (com valores padrão)

```env
RATE_LIMIT_REQUESTS=10       # Máximo de requisições
RATE_LIMIT_PERIOD=1          # Período em segundos
THROTTLE_MIN_MS=0            # Delay mínimo (0 = desabilitado)
THROTTLE_MAX_MS=0            # Delay máximo
```

---

## 🧪 Testar Deploy

### 1. Health Check

```bash
curl https://sua-url.com/health | jq '.'
```

Resposta esperada:
```json
{
  "status": "ok",
  "database": {
    "status": "connected"
  },
  "configuration": {
    "rate_limiting": {...},
    "throttling": {...}
  }
}
```

### 2. Testar Endpoints

```bash
# GET
curl https://sua-url.com/api/get

# POST
curl -X POST https://sua-url.com/api/post \
  -H "Content-Type: application/json" \
  -d '{"test":"production"}'

# Banco
curl -X POST https://sua-url.com/api/db/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"Mensagem de produção"}'
```

---

## 🐛 Troubleshooting

### Erro: "Database connection refused"

**Causa**: Configuração incorreta do host do banco.

**Solução**:
- Verifique se `DB_HOST` está correto
- Use host interno da plataforma (não externo)
- Exemplos:
  - Dokploy: `nome-do-service:5432`
  - Railway: Use variáveis do Postgres
  - Render: Use Internal Database URL

### Erro: "Port already in use"

**Causa**: Plataforma espera porta diferente.

**Solução**:
- Render: Sempre use `PORT=10000` ou a variável `$PORT`
- Railway/Fly.io: Use a porta que você configurou
- Dokploy: Configurável no painel

### Erro: "Build failed"

**Causa**: Nixpacks não encontrou os arquivos Go.

**Solução**:
- Verifique se `nixpacks.toml` está na raiz
- Confirme que `server/main.go` existe
- Veja os logs de build para detalhes

### Health Check retorna "degraded"

**Causa**: Banco de dados não está acessível.

**Solução**:
1. Verifique logs: `flyctl logs` / `railway logs` / etc.
2. Confirme variáveis de ambiente
3. Teste conexão com banco manualmente
4. Verifique se o banco está rodando

---

## 📊 Monitoramento em Produção

### Logs

```bash
# Railway
railway logs --follow

# Fly.io
flyctl logs

# Render
# Via dashboard em "Logs"

# Dokploy
# Via painel "Logs"
```

### Métricas

Use o endpoint `/health` para monitoramento:

```bash
# Script de monitoramento
while true; do
  STATUS=$(curl -s https://sua-url.com/health | jq -r '.status')
  echo "$(date): Status = $STATUS"
  sleep 60
done
```

### Alertas

Configure alertas na plataforma:
- **Railway**: Notifications → Configure alerts
- **Fly.io**: `flyctl monitor`
- **Render**: Dashboard → Alerts
- **Dokploy**: Monitoring → Alerts

---

## 🔒 Segurança em Produção

### ⚠️ Checklist de Segurança

- [ ] Senhas do banco são fortes e únicas
- [ ] Variáveis de ambiente estão seguras (não commitadas)
- [ ] Rate limiting configurado adequadamente
- [ ] CORS configurado (se necessário)
- [ ] HTTPS habilitado (automático nas plataformas)
- [ ] Logs não expõem dados sensíveis
- [ ] Database backups configurados

### Recomendações

1. **Não commite** `.env.production` com valores reais
2. Use **secrets management** da plataforma
3. Configure **database backups** automáticos
4. Monitore **health checks** regularmente
5. Configure **rate limiting** apropriado para sua necessidade

---

## 💰 Custos Estimados

### Free Tiers

| Plataforma | API | PostgreSQL | Limitações |
|------------|-----|------------|------------|
| **Railway** | $5/mês (500h) | Incluído | Hibernação após 6h inativo |
| **Render** | Free | $7/mês | App suspende após 15min |
| **Fly.io** | Free (256MB) | Free (1GB) | 3 apps gratuitos |
| **Dokploy** | Self-hosted | Self-hosted | Custo do VPS apenas |

### Produção Recomendada

- **Dokploy** (Self-hosted): ~$5-10/mês (VPS)
- **Railway**: ~$10-20/mês
- **Render**: ~$15-25/mês
- **Fly.io**: ~$10-15/mês

---

## 📚 Recursos Adicionais

- [Nixpacks Documentation](https://nixpacks.com/)
- [Dokploy Docs](https://dokploy.com/docs)
- [Railway Docs](https://docs.railway.app/)
- [Render Docs](https://render.com/docs)
- [Fly.io Docs](https://fly.io/docs/)

---

Voltar para o [README principal](README.md)

