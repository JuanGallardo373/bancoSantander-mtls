#!/usr/bin/env python3
"""
Script de Atacante - Intenta conectar con certificado autofirmado
Simula un ataque MITM o suplantación de identidad contra el servidor Santander
"""

import requests
import json
import time
import sys
from datetime import datetime
from requests.adapters import HTTPAdapter
from urllib3.util.ssl_ import create_urllib3_context
import warnings

# Suprimir advertencias de SSL
warnings.filterwarnings('ignore', message='Unverified HTTPS request')

class AttackerMTLSAdapter(HTTPAdapter):
    """Adaptador para intentar conexión con certificado autofirmado"""
    
    def __init__(self, cert_file, key_file, **kwargs):
        self.cert_file = cert_file
        self.key_file = key_file
        super().__init__(**kwargs)
    
    def init_poolmanager(self, *args, **kwargs):
        ctx = create_urllib3_context()
        # Intentar cargar certificado autofirmado
        try:
            ctx.load_cert_chain(self.cert_file, self.key_file)
        except Exception as e:
            print(f"⚠️  Error cargando certificado autofirmado: {e}")
        kwargs['ssl_context'] = ctx
        return super().init_poolmanager(*args, **kwargs)

def create_attacker_session(cert_path, key_path):
    """Crea una sesión con certificado autofirmado"""
    session = requests.Session()
    adapter = AttackerMTLSAdapter(
        cert_file=cert_path,
        key_file=key_path,
        pool_connections=10,
        pool_maxsize=10
    )
    session.mount('https://', adapter)
    return session

def create_malicious_transfer(amount, target_account):
    """Crea una solicitud de transferencia maliciosa"""
    return {
        "source_bank": "ATTACKER_BANK",
        "source_account": "ATTACKER-ACC-001",
        "destination_bank": "Santander",
        "destination_account": target_account,
        "amount": amount,
        "currency": "USD"
    }

def attempt_transfer(session, server_url, transfer_data, attempt_num, timeout=10):
    """Intenta enviar una transferencia maliciosa"""
    try:
        print(f"\n[Intento {attempt_num}] Enviando solicitud maliciosa...")
        print(f"   → Certificado: {transfer_data['source_bank']}")
        print(f"   → Monto: ${transfer_data['amount']:.2f}")
        print(f"   → Cuenta destino: {transfer_data['destination_account']}")
        
        response = session.post(
            f"{server_url}/transfer",
            json=transfer_data,
            timeout=timeout,
            verify=False  # No verificar certificado del servidor (ataque MITM)
        )
        
        return response
    
    except requests.exceptions.SSLError as e:
        return None, str(e)
    except requests.exceptions.RequestException as e:
        return None, str(e)

def log_attack_attempt(timestamp, attempt_num, status, error_message, amount):
    """Registra un intento de ataque"""
    log_entry = {
        "timestamp": timestamp,
        "attacker": "malicious_actor",
        "attempt_number": attempt_num,
        "status": status,
        "error": error_message,
        "amount_attempted": amount,
        "certificate_type": "self-signed"
    }
    print("\n📋 Log de intento de ataque:")
    print(json.dumps(log_entry, indent=2, default=str))

def print_header():
    """Imprime banner de atacante"""
    header = """
╔══════════════════════════════════════════════════════════════╗
║                   🔴 SIMULATED ATTACKER 🔴                  ║
║                                                              ║
║  Intentando acceder al servidor Santander con              ║
║  certificado autofirmado (NO AUTORIZADO)                   ║
║                                                              ║
║  ⚠️  Este script es solo para fines educativos             ║
║  ⚠️  Se espera que FALLE y sea DETECTADO                   ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
"""
    print(header)

def main():
    print_header()
    
    # Configuración del atacante
    SERVER_URL = "https://localhost:8443"
    ATTACKER_CERT = "./certs/attacker-self-signed.pem"
    ATTACKER_KEY = "./certs/attacker-self-signed.key"
    
    print("🎯 Objetivo: Banco Santander")
    print(f"📍 Servidor: {SERVER_URL}")
    print(f"🔓 Certificado (autofirmado): {ATTACKER_CERT}")
    print()
    
    print("=" * 62)
    print("🔍 Verificación de disponibilidad del servidor...")
    print("=" * 62)
    
    try:
        # Intentar health check (sin mTLS)
        response = requests.get(f"{SERVER_URL}/health", verify=False, timeout=5)
        if response.status_code == 200:
            print("✓ Servidor disponible")
            print(f"  Respuesta: {response.json()}")
        else:
            print(f"⚠️  Servidor responde con código {response.status_code}")
    except Exception as e:
        print(f"❌ Servidor no disponible: {e}")
        print("   Asegúrate de que el servidor Santander está corriendo:")
        print("   $ cd servidor-banco && go run main.go")
        return
    
    print()
    print("=" * 62)
    print("🚨 INICIANDO ATAQUES CON CERTIFICADO AUTOFIRMADO...")
    print("=" * 62)
    
    try:
        # Crear sesión con certificado autofirmado
        print("\n🔐 Configurando cliente con certificado autofirmado...")
        session = create_attacker_session(ATTACKER_CERT, ATTACKER_KEY)
        print("✓ Sesión configurada (certificado NO verificado por CA)")
        
        # Intentos de transferencia maliciosa
        attack_scenarios = [
            {
                "description": "Transferencia grande no autorizada",
                "amount": 999999.99,
                "target": "ACC-VICTIM-001"
            },
            {
                "description": "Suplantación de BBVA",
                "amount": 50000.00,
                "target": "ACC-SANTANDER-EXECUTIVE"
            },
            {
                "description": "Extracción de fondos",
                "amount": 100000.00,
                "target": "ACC-ATTACKER-MONEY-LAUNDRY"
            },
        ]
        
        successful_attacks = 0
        
        for i, scenario in enumerate(attack_scenarios, 1):
            timestamp = datetime.now().isoformat()
            
            # Crear solicitud maliciosa
            malicious_transfer = create_malicious_transfer(
                amount=scenario["amount"],
                target_account=scenario["target"]
            )
            
            print(f"\n{'─' * 62}")
            print(f"📤 Escenario {i}: {scenario['description']}")
            print(f"{'─' * 62}")
            
            # Intentar enviar
            response = attempt_transfer(session, SERVER_URL, malicious_transfer, i)
            
            if isinstance(response, tuple):
                response, error_msg = response
            
            if response:
                if response.status_code == 200:
                    print(f"\n🚨 ¡¡ATAQUE EXITOSO!! (No debería suceder)")
                    print(f"   Respuesta del servidor: {response.json()}")
                    log_attack_attempt(timestamp, i, "SUCCESSFUL", "Certificado aceptado", scenario["amount"])
                    successful_attacks += 1
                else:
                    print(f"\n❌ Ataque rechazado (código {response.status_code})")
                    print(f"   Respuesta: {response.text}")
                    log_attack_attempt(timestamp, i, "REJECTED", response.text, scenario["amount"])
            else:
                print(f"\n❌ Conexión rechazada (SSL/mTLS)")
                print(f"   Error: {error_msg}")
                log_attack_attempt(timestamp, i, "SSL_ERROR", error_msg, scenario["amount"])
            
            # Esperar entre intentos
            if i < len(attack_scenarios):
                time.sleep(1)
        
        # Resumen
        print()
        print("=" * 62)
        print("📊 RESUMEN DE INTENTOS DE ATAQUE")
        print("=" * 62)
        print(f"Total de intentos: {len(attack_scenarios)}")
        print(f"Ataques exitosos: {successful_attacks}")
        print(f"Ataques bloqueados: {len(attack_scenarios) - successful_attacks}")
        
        if successful_attacks == 0:
            print("\n✅ RESULTADO ESPERADO: Todos los ataques fueron bloqueados")
            print("   El servidor Santander rechazó correctamente el certificado autofirmado")
            print("\n📋 Los intentos de ataque han sido registrados en:")
            print("   ../logs/anomalies.jsonl")
            print("\n🤖 Ejecuta el analizador LLM para ver el análisis de seguridad:")
            print("   python3 analista-ia/llm_analyzer.py")
        else:
            print(f"\n🚨 FALLO DE SEGURIDAD: {successful_attacks} ataques fueron exitosos")
            print("   ¡El servidor debe rechazar certificados autofirmados!")
        
        print()
        print("=" * 62)
        
    except Exception as e:
        print(f"\n❌ Error durante el ataque: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()



