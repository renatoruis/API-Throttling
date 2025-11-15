.PHONY: help build up down restart logs test clean

help:
	@echo "API Throttling - Comandos Disponíveis"
	@echo ""
	@echo "  make build     - Constrói as imagens Docker"
	@echo "  make up        - Inicia os serviços"
	@echo "  make down      - Para e remove os serviços"
	@echo "  make restart   - Reinicia a API"
	@echo "  make logs      - Mostra logs da API"
	@echo "  make test      - Executa testes"
	@echo "  make clean     - Remove volumes e imagens"
	@echo ""
	@echo "Cenários de teste:"
	@echo "  make scenario-fast      - API rápida (sem limitações)"
	@echo "  make scenario-normal    - API normal (pequeno delay)"
	@echo "  make scenario-slow      - API lenta (servidor sobrecarregado)"
	@echo "  make scenario-strict    - API muito restritiva"

build:
	docker-compose build

up:
	docker-compose up -d
	@echo "Aguardando API iniciar..."
	@sleep 5
	@echo "API disponível em http://localhost:8888"

down:
	docker-compose down

restart:
	docker-compose restart api
	@echo "API reiniciada"

logs:
	docker-compose logs -f api

test:
	@cd client-tests && ./test-api.sh

clean:
	docker-compose down -v
	docker-compose down --rmi local

# Cenários pré-configurados
scenario-fast:
	@echo "Configurando: API Rápida (sem limitações)"
	@sed -i.bak 's/RATE_LIMIT_REQUESTS=.*/RATE_LIMIT_REQUESTS=1000/' docker-compose.yml
	@sed -i.bak 's/THROTTLE_MIN_MS=.*/THROTTLE_MIN_MS=0/' docker-compose.yml
	@sed -i.bak 's/THROTTLE_MAX_MS=.*/THROTTLE_MAX_MS=0/' docker-compose.yml
	@rm docker-compose.yml.bak
	@make restart
	@echo "✓ Cenário configurado: API Rápida"

scenario-normal:
	@echo "Configurando: API Normal (pequeno delay)"
	@sed -i.bak 's/RATE_LIMIT_REQUESTS=.*/RATE_LIMIT_REQUESTS=50/' docker-compose.yml
	@sed -i.bak 's/THROTTLE_MIN_MS=.*/THROTTLE_MIN_MS=50/' docker-compose.yml
	@sed -i.bak 's/THROTTLE_MAX_MS=.*/THROTTLE_MAX_MS=150/' docker-compose.yml
	@rm docker-compose.yml.bak
	@make restart
	@echo "✓ Cenário configurado: API Normal"

scenario-slow:
	@echo "Configurando: API Lenta (servidor sobrecarregado)"
	@sed -i.bak 's/RATE_LIMIT_REQUESTS=.*/RATE_LIMIT_REQUESTS=10/' docker-compose.yml
	@sed -i.bak 's/THROTTLE_MIN_MS=.*/THROTTLE_MIN_MS=500/' docker-compose.yml
	@sed -i.bak 's/THROTTLE_MAX_MS=.*/THROTTLE_MAX_MS=2000/' docker-compose.yml
	@rm docker-compose.yml.bak
	@make restart
	@echo "✓ Cenário configurado: API Lenta"

scenario-strict:
	@echo "Configurando: API Muito Restritiva"
	@sed -i.bak 's/RATE_LIMIT_REQUESTS=.*/RATE_LIMIT_REQUESTS=5/' docker-compose.yml
	@sed -i.bak 's/THROTTLE_MIN_MS=.*/THROTTLE_MIN_MS=1000/' docker-compose.yml
	@sed -i.bak 's/THROTTLE_MAX_MS=.*/THROTTLE_MAX_MS=3000/' docker-compose.yml
	@rm docker-compose.yml.bak
	@make restart
	@echo "✓ Cenário configurado: API Muito Restritiva"

# Atalhos úteis
status:
	@docker-compose ps

db:
	docker-compose exec postgres psql -U postgres -d apidb

health:
	@echo "🏥 Health Check da API"
	@echo ""
	@curl -s http://localhost:8888/health | jq '.'
	@echo ""
	@echo "Status: $$(curl -s http://localhost:8888/health | jq -r '.status')"
	@echo "Database: $$(curl -s http://localhost:8888/health | jq -r '.database.status')"


