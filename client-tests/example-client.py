#!/usr/bin/env python3
"""
Cliente de exemplo para testar a API com Throttling e Rate Limiting

Instalação:
    pip3 install -r requirements.txt

Uso:
    python3 example-client.py
"""

import requests
import time
import statistics
from datetime import datetime
from typing import List

API_URL = "http://localhost:8888"


def print_header(title: str):
    """Imprime um cabeçalho formatado"""
    print("\n" + "=" * 60)
    print(f"  {title}")
    print("=" * 60 + "\n")


def test_throttling(num_requests: int = 5):
    """Testa o throttling medindo latência"""
    print_header("Teste de Throttling (Latência)")
    print(f"Enviando {num_requests} requisições e medindo tempo de resposta...\n")
    
    latencies: List[float] = []
    
    for i in range(1, num_requests + 1):
        start = time.time()
        try:
            response = requests.get(f"{API_URL}/api/get", timeout=10)
            end = time.time()
            latency_ms = (end - start) * 1000
            latencies.append(latency_ms)
            
            status_icon = "✓" if response.status_code == 200 else "✗"
            print(f"Request {i:2d}: {status_icon} {latency_ms:6.0f}ms (HTTP {response.status_code})")
        except Exception as e:
            print(f"Request {i:2d}: ✗ Erro: {e}")
    
    if latencies:
        print(f"\nEstatísticas:")
        print(f"  Média:   {statistics.mean(latencies):6.0f}ms")
        print(f"  Mínima:  {min(latencies):6.0f}ms")
        print(f"  Máxima:  {max(latencies):6.0f}ms")
        print(f"  Mediana: {statistics.median(latencies):6.0f}ms")


def test_rate_limiting(num_requests: int = 20):
    """Testa o rate limiting"""
    print_header("Teste de Rate Limiting")
    print(f"Enviando {num_requests} requisições rápidas...\n")
    
    successful = 0
    rate_limited = 0
    
    for i in range(1, num_requests + 1):
        try:
            response = requests.get(f"{API_URL}/api/get", timeout=10)
            
            if response.status_code == 200:
                print(f"Request {i:2d}: ✓ 200 OK")
                successful += 1
            elif response.status_code == 429:
                print(f"Request {i:2d}: ✗ 429 Rate Limited")
                rate_limited += 1
            else:
                print(f"Request {i:2d}: ? {response.status_code}")
        except Exception as e:
            print(f"Request {i:2d}: ✗ Erro: {e}")
    
    print(f"\nResultado:")
    print(f"  Bem-sucedidas:      {successful:3d} ({successful/num_requests*100:.0f}%)")
    print(f"  Rate Limited (429): {rate_limited:3d} ({rate_limited/num_requests*100:.0f}%)")


def test_post_endpoint():
    """Testa o endpoint POST"""
    print_header("Teste de POST com Payload")
    
    payload = {
        "test": "python_client",
        "timestamp": datetime.now().isoformat(),
        "value": 42
    }
    
    try:
        response = requests.post(
            f"{API_URL}/api/post",
            json=payload,
            timeout=10
        )
        
        if response.status_code == 200:
            print("✓ POST bem-sucedido\n")
            print("Resposta:")
            import json
            print(json.dumps(response.json(), indent=2))
        else:
            print(f"✗ POST falhou (HTTP {response.status_code})")
    except Exception as e:
        print(f"✗ Erro: {e}")


def test_database():
    """Testa endpoints do banco de dados"""
    print_header("Teste de Banco de Dados")
    
    # Salvar mensagem
    print("1. Salvando mensagem no banco...")
    message = {
        "content": f"Mensagem Python em {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"
    }
    
    try:
        response = requests.post(
            f"{API_URL}/api/db/messages",
            json=message,
            timeout=10
        )
        
        if response.status_code == 201:
            print("   ✓ Mensagem salva com sucesso")
            data = response.json().get('data', {})
            print(f"   ID: {data.get('id')}, Content: {data.get('content')}")
        else:
            print(f"   ✗ Falha ao salvar (HTTP {response.status_code})")
    except Exception as e:
        print(f"   ✗ Erro: {e}")
    
    # Listar mensagens
    print("\n2. Listando mensagens do banco...")
    try:
        response = requests.get(f"{API_URL}/api/db/messages", timeout=10)
        
        if response.status_code == 200:
            data = response.json()
            count = data.get('count', 0)
            print(f"   ✓ Total de mensagens: {count}")
            
            messages = data.get('messages', [])
            if messages:
                print("\n   Últimas 5 mensagens:")
                for msg in messages[:5]:
                    print(f"   - [{msg.get('id')}] {msg.get('content')[:50]}...")
        else:
            print(f"   ✗ Falha ao listar (HTTP {response.status_code})")
    except Exception as e:
        print(f"   ✗ Erro: {e}")


def test_concurrent_requests(num_requests: int = 10):
    """Testa requisições concorrentes"""
    print_header("Teste de Requisições Concorrentes")
    print(f"Enviando {num_requests} requisições em paralelo...\n")
    
    import concurrent.futures
    
    def make_request(i: int) -> tuple:
        start = time.time()
        try:
            response = requests.get(f"{API_URL}/api/get", timeout=10)
            latency = (time.time() - start) * 1000
            return (i, response.status_code, latency)
        except Exception as e:
            return (i, 0, 0)
    
    with concurrent.futures.ThreadPoolExecutor(max_workers=num_requests) as executor:
        futures = [executor.submit(make_request, i) for i in range(1, num_requests + 1)]
        results = [f.result() for f in concurrent.futures.as_completed(futures)]
    
    # Ordenar por número da requisição
    results.sort(key=lambda x: x[0])
    
    successful = 0
    rate_limited = 0
    
    for i, status, latency in results:
        if status == 200:
            print(f"Request {i:2d}: ✓ 200 OK ({latency:.0f}ms)")
            successful += 1
        elif status == 429:
            print(f"Request {i:2d}: ✗ 429 Rate Limited")
            rate_limited += 1
        else:
            print(f"Request {i:2d}: ✗ Erro")
    
    print(f"\nResultado:")
    print(f"  Bem-sucedidas: {successful}")
    print(f"  Rate Limited:  {rate_limited}")


def main():
    """Executa todos os testes"""
    print("=" * 60)
    print("  API Throttling - Cliente de Teste Python")
    print("=" * 60)
    
    # Verificar se a API está rodando
    try:
        response = requests.get(f"{API_URL}/health", timeout=5)
        if response.status_code in [200, 503]:
            health_data = response.json()
            status = health_data.get('status', 'unknown')
            db_status = health_data.get('database', {}).get('status', 'unknown')
            
            if status == 'ok':
                print(f"\n✓ API está online (DB: {db_status})")
            elif status == 'degraded':
                print(f"\n⚠ API está online mas degradada (DB: {db_status})")
            
            # Mostrar configuração
            config = health_data.get('configuration', {})
            print("\nConfiguração ativa:")
            print(f"  Rate Limiting: {config.get('rate_limiting', {}).get('requests', '?')} req/{config.get('rate_limiting', {}).get('period_seconds', '?')}s")
            
            throttling = config.get('throttling', {})
            if throttling.get('enabled'):
                print(f"  Throttling: {throttling.get('min_ms', '?')}-{throttling.get('max_ms', '?')}ms")
            else:
                print("  Throttling: desabilitado")
        else:
            print(f"\n✗ API retornou status inesperado: {response.status_code}")
            return
    except Exception as e:
        print(f"\n✗ API não está respondendo: {e}")
        print("\nExecute: docker-compose up -d")
        return
    
    # Executar testes
    test_throttling(5)
    time.sleep(1)  # Aguardar para não bater no rate limit entre testes
    
    test_rate_limiting(20)
    time.sleep(1)
    
    test_post_endpoint()
    time.sleep(1)
    
    test_database()
    time.sleep(1)
    
    test_concurrent_requests(10)
    
    print("\n" + "=" * 60)
    print("  Testes Concluídos!")
    print("=" * 60 + "\n")


if __name__ == "__main__":
    main()


