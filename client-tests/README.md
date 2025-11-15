# Clientes de Teste

Scripts prontos para testar a API de Throttling.

## 🐚 Script Bash

### Uso

```bash
./test-api.sh
```

### O que faz

- Verifica se a API está online
- Testa throttling (mede latência)
- Testa rate limiting (identifica rejeições HTTP 429)
- Testa POST com payload
- Testa integração com banco de dados

## 🐍 Cliente Python

### Instalação

```bash
pip3 install -r requirements.txt
```

### Uso

```bash
python3 example-client.py
```

### O que faz

- Teste de throttling com estatísticas (média, mínima, máxima, mediana)
- Teste de rate limiting com contagem
- Teste de POST com payload
- Teste de banco de dados (salvar e listar)
- Teste de requisições concorrentes

### Recursos

- Medição de latência em milissegundos
- Estatísticas detalhadas
- Tratamento de erros
- Output formatado e colorido
- Exemplos de retry com exponential backoff

## 🔧 Configuração

Ambos os scripts usam por padrão:

```bash
API_URL="http://localhost:8888"
```

Para alterar:

```bash
# Bash
export API_URL="http://localhost:9999"
./test-api.sh

# Python (editar diretamente no arquivo)
API_URL = "http://seu-servidor:porta"
```

## 📝 Exemplos Adicionais

### Teste simples com curl

```bash
# Health check
curl http://localhost:8888/health

# GET
curl http://localhost:8888/api/get

# POST
curl -X POST http://localhost:8888/api/post \
  -H "Content-Type: application/json" \
  -d '{"test":"value"}'
```

### Teste de carga com loop

```bash
# 50 requisições sequenciais
for i in {1..50}; do
  curl -s http://localhost:8888/api/get > /dev/null
  echo "Request $i done"
done
```

## 🎯 Casos de Uso

Use estes clientes para:

1. **Validar configuração**: Verificar se throttling e rate limiting estão funcionando
2. **Testes automatizados**: Integrar em pipelines CI/CD
3. **Benchmarks**: Medir performance sob diferentes configurações
4. **Demonstrações**: Mostrar comportamento de APIs com limitações
5. **Aprendizado**: Estudar como implementar clientes resilientes

---

Voltar para o [README principal](../README.md)

