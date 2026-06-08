import os
import time
import json
from datetime import datetime
from ollama import Client

# 1. Configuración de rutas y cliente local
# Asumimos que ejecutas este script parado dentro de la carpeta 'analista-ia/'
LOG_BANCO_PATH = "../servidor-banco/logs/anomalies.log"
LOG_ANALYSIS_PATH = "logs/analisys.log"
MODEL_NAME = "analista-mtls"

ollama_client = Client(host="http://localhost:11434")

def analizar_con_llm(linea_log):
    """Envía la línea de log al LLM local y procesa la respuesta JSON"""
    try:
        response = ollama_client.generate(
            model=MODEL_NAME,
            prompt=f"Analiza la siguiente línea de log del servidor bancario: {linea_log}"
        )
        return json.loads(response['response'].strip())
    except json.JSONDecodeError:
        return {"estado": "ERROR_PARSING", "raw_output": response.get('response', '')}
    except Exception as e:
        return {"estado": "FALLO_API", "detalle": str(e)}

def guardar_anomalia(alerta_json):
    """Guarda el JSON de la alerta en el archivo anomalies.log con un Timestamp"""
    # Asegurarse de que la carpeta logs/ exista dentro de analista-ia/
    os.makedirs(os.path.dirname(LOG_ANOMALIAS_PATH), exist_ok=True)
    
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    registro_persistente = {
        "timestamp": timestamp,
        "evento": alerta_json
    }
    
    # Escribir en modo 'append' (a) para no borrar el historial
    with open(LOG_ANOMALIAS_PATH, 'a', encoding='utf-8') as archivo:
        archivo.write(json.dumps(registro_persistente, ensure_ascii=False) + "\n")

def notificar_administrador(alerta_json):
    """Simula una alerta crítica por consola (ej. SMS, Email, PagerDuty)"""
    print("\n" + "!" * 70)
    print("🚨 [ALERTA CRÍTICA DE SEGURIDAD] 🚨")
    print("Notificando al equipo de SOC / Administrador del Banco Central...")
    print(f"➤ Categoría: {alerta_json.get('categoria', 'Desconocida')}")
    print(f"➤ Origen IP: {alerta_json.get('origen_ip', 'Desconocida')}")
    print(f"➤ Acción Recomendada: {alerta_json.get('accion_recomendada', 'Revisar logs')}")
    print("!" * 70 + "\n")

def monitorear_logs():
    """Simula un 'tail -f' sobre el archivo de log del sistema operativo"""
    print("[*] Iniciando Sistema de Prevención de Intrusiones (IPS) impulsado por IA...")
    print(f"[*] Monitoreando activamente: {LOG_BANCO_PATH}")
    print(f"[*] Guardando historial de anomalías en: {LOG_ANOMALIAS_PATH}")
    print("-" * 70)

    # Crear el archivo del banco si no existe para evitar cuelgues
    if not os.path.exists(LOG_BANCO_PATH):
        os.makedirs(os.path.dirname(LOG_BANCO_PATH), exist_ok=True)
        with open(LOG_BANCO_PATH, 'w') as f:
            f.write("")

    with open(LOG_BANCO_PATH, 'r') as archivo:
        archivo.seek(0, os.SEEK_END)
        
        while True:
            linea = archivo.readline()
            if not linea:
                time.sleep(0.5)
                continue
                
            # Filtrar eventos de TLS
            if "TLS handshake error" in linea or "failed to verify" in linea:
                print(f"\n[!] Anomalía detectada en Capa de Transporte (TLS):\n    {linea.strip()}")
                print("[*] Solicitando análisis heurístico a Llama 3...")
                
                alerta = analizar_con_llm(linea)
                
                # Mostrar el JSON bonito en pantalla
                print(json.dumps(alerta, indent=4, ensure_ascii=False))
                
                # --- NUEVA LÓGICA DE ALMACENAMIENTO Y ALERTAS ---
                if alerta.get("estado") == "ALERTA":
                    # 1. Guardar siempre en el historial
                    guardar_anomalia(alerta)
                    print(f"[*] Registro guardado exitosamente en {LOG_ANOMALIAS_PATH}")
                    
                    # 2. Filtrar si es lo suficientemente grave para despertar al administrador
                    categorias_criticas = ["INTENTO DE SUPLANTACIÓN", "INCOMPATIBILIDAD DE ROL"]
                    if alerta.get("categoria") in categorias_criticas:
                        notificar_administrador(alerta)
                # ------------------------------------------------

                print("-" * 70)

if __name__ == "__main__":
    try:
        monitorear_logs()
    except KeyboardInterrupt:
        print("\n[-] Agente analista detenido por el usuario.")
