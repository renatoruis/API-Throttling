#!/bin/bash

# Script de teste para demonstrar Rate Limiting e Throttling

API_URL="${API_URL:-http://localhost:8888}"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "  API Throttling - Script de Teste"
echo "=========================================="
echo ""

# Check if API is running
echo -n "Verificando se a API está rodando... "
HEALTH_RESPONSE=$(curl -s "${API_URL}/health")
if [ $? -eq 0 ]; then
    STATUS=$(echo "$HEALTH_RESPONSE" | jq -r '.status' 2>/dev/null || echo "unknown")
    DB_STATUS=$(echo "$HEALTH_RESPONSE" | jq -r '.database.status' 2>/dev/null || echo "unknown")
    
    if [ "$STATUS" == "ok" ]; then
        echo -e "${GREEN}✓ API está online (DB: $DB_STATUS)${NC}"
    elif [ "$STATUS" == "degraded" ]; then
        echo -e "${YELLOW}⚠ API online mas degradada (DB: $DB_STATUS)${NC}"
    else
        echo -e "${YELLOW}? API respondeu mas status desconhecido${NC}"
    fi
    
    # Mostrar configuração
    echo ""
    echo "Configuração ativa:"
    echo "$HEALTH_RESPONSE" | jq '.configuration' 2>/dev/null || echo "Não foi possível ler configuração"
    echo ""
else
    echo -e "${RED}✗ API não está respondendo${NC}"
    echo "Execute: docker-compose up -d"
    exit 1
fi

echo ""
echo "=========================================="
echo "  Teste 1: Throttling (Latência)"
echo "=========================================="
echo "Medindo tempo de resposta de 3 requisições..."
echo ""

for i in {1..3}; do
    echo -n "Request $i: "
    START=$(date +%s%3N)
    curl -s "${API_URL}/api/get" > /dev/null
    END=$(date +%s%3N)
    DURATION=$((END - START))
    echo -e "${YELLOW}${DURATION}ms${NC}"
done

echo ""
echo "=========================================="
echo "  Teste 2: Rate Limiting"
echo "=========================================="
echo "Enviando 15 requisições rápidas..."
echo "Algumas devem ser rejeitadas (HTTP 429)"
echo ""

SUCCESS=0
REJECTED=0

for i in {1..15}; do
    STATUS=$(curl -s -w "%{http_code}" -o /dev/null "${API_URL}/api/get")
    if [ "$STATUS" == "200" ]; then
        echo -e "Request $i: ${GREEN}✓ 200 OK${NC}"
        SUCCESS=$((SUCCESS + 1))
    else
        echo -e "Request $i: ${RED}✗ $STATUS Rate Limited${NC}"
        REJECTED=$((REJECTED + 1))
    fi
done

echo ""
echo "Resultado:"
echo -e "  ${GREEN}Aceitas: $SUCCESS${NC}"
echo -e "  ${RED}Rejeitadas: $REJECTED${NC}"

echo ""
echo "=========================================="
echo "  Teste 3: POST com Payload"
echo "=========================================="
echo ""

RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"test":"payload","value":123}' \
    "${API_URL}/api/post")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_CODE/d')

if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}✓ POST bem-sucedido${NC}"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
    echo -e "${RED}✗ POST falhou (HTTP $HTTP_CODE)${NC}"
fi

echo ""
echo "=========================================="
echo "  Teste 4: Banco de Dados"
echo "=========================================="
echo ""

# Salvar mensagem
echo "Salvando mensagem no banco..."
RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "{\"content\":\"Teste em $(date +'%Y-%m-%d %H:%M:%S')\"}" \
    "${API_URL}/api/db/messages")

if echo "$RESPONSE" | grep -q "saved successfully"; then
    echo -e "${GREEN}✓ Mensagem salva com sucesso${NC}"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
else
    echo -e "${RED}✗ Falha ao salvar mensagem${NC}"
fi

echo ""
echo "Listando mensagens do banco..."
MESSAGES=$(curl -s "${API_URL}/api/db/messages")
COUNT=$(echo "$MESSAGES" | jq -r '.count' 2>/dev/null || echo "0")

echo -e "${GREEN}Total de mensagens: $COUNT${NC}"
echo "$MESSAGES" | jq '.' 2>/dev/null || echo "$MESSAGES"

echo ""
echo "=========================================="
echo "  Testes Concluídos!"
echo "=========================================="
echo ""
echo "Para modificar a configuração:"
echo "  1. Edite docker-compose.yml"
echo "  2. Execute: docker-compose restart api"
echo ""
echo "Configuração atual (veja nos logs):"
docker-compose logs api 2>/dev/null | grep -E "(Rate limiter|Throttling)" | tail -2 || echo "Execute: docker-compose logs api | grep -E 'Rate|Throttling'"
echo ""


