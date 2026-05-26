#!/usr/bin/env python3
"""
Analizador de Anomalías mTLS usando LLM local (Ollama)
Detecta patrones de ataque y anomalías en los logs de handshakes mTLS
"""

import json
import requests
import sys
from datetime import datetime, timedelta
from pathlib import Path
import time

class MTLSAnomalyAnalyzer:
    """Analizador de anomalías mTLS con LLM local"""
    
    def __init__(self, ollama_url="http://localhost:11434", model="llama2"):
        """
        Inicializa el analizador
        
        Args:
            ollama_url: URL del servidor Ollama
            model: Modelo LLM a usar
        """
        self.ollama_url = ollama_url
        self.model = model
        self.log_file = "../logs/anomalies.jsonl"
        self.analysis_file = "../logs/analysis.jsonl"
    
    def check_ollama_availability(self):
        """Verifica si Ollama está disponible"""
        try:
            response = requests.get(f"{self.ollama_url}/api/tags", timeout=5)
            return response.status_code == 200
        except requests.RequestException:
            return False
    
    def read_recent_logs(self, minutes=60):
        """Lee logs recientes desde el archivo JSONL"""
        anomalies = []
        
        if not Path(self.log_file).exists():
            print(f"⚠️  Archivo de logs no encontrado: {self.log_file}")
            return anomalies
        
        cutoff_time = datetime.now() - timedelta(minutes=minutes)
        
        try:
            with open(self.log_file, 'r') as f:
                for line in f:
                    if line.strip():
                        try:
                            log_entry = json.loads(line)
                            timestamp = datetime.fromisoformat(log_entry.get('timestamp', ''))
                            if timestamp > cutoff_time:
                                anomalies.append(log_entry)
                        except json.JSONDecodeError:
                            continue
        except Exception as e:
            print(f"❌ Error leyendo logs: {e}")
        
        return anomalies
    
    def format_anomalies_for_analysis(self, anomalies):
        """Formatea las anomalías para análisis por LLM"""
        if not anomalies:
            return "Sin anomalías detectadas en los logs recientes."
        
        formatted = "ANOMALÍAS DETECTADAS EN LOGS mTLS:\n\n"
        for i, anomaly in enumerate(anomalies, 1):
            formatted += f"{i}. Timestamp: {anomaly.get('timestamp')}\n"
            formatted += f"   Cliente: {anomaly.get('client_name', 'Desconocido')}\n"
            formatted += f"   IP: {anomaly.get('client_ip', 'N/A')}\n"
            formatted += f"   Tipo: {anomaly.get('event_type', 'N/A')}\n"
            formatted += f"   Certificado Expirado: {anomaly.get('is_expired', False)}\n"
            formatted += f"   Autofirmado: {anomaly.get('is_self_signed', False)}\n"
            formatted += f"   Mensaje: {anomaly.get('error_message', 'N/A')}\n\n"
        
        return formatted
    
    def analyze_with_llm(self, anomalies_text):
        """Analiza anomalías usando LLM local (Ollama)"""
        prompt = f"""Analiza las siguientes anomalías de seguridad mTLS en un servidor bancario. 
Identifica patrones de ataque, valida si son intentos maliciosos y proporciona recomendaciones de seguridad.

{anomalies_text}

Por favor proporciona:
1. Clasificación de riesgo (CRÍTICO, ALTO, MEDIO, BAJO)
2. Análisis de cada anomalía
3. Patrones detectados
4. Recomendaciones de acción inmediata
5. Si se debe notificar al administrador

Sé conciso y enfocado en la seguridad."""
        
        try:
            response = requests.post(
                f"{self.ollama_url}/api/generate",
                json={
                    "model": self.model,
                    "prompt": prompt,
                    "stream": False,
                    "temperature": 0.3
                },
                timeout=60
            )
            
            if response.status_code == 200:
                return response.json().get('response', '')
            else:
                return f"Error en LLM: {response.status_code} - {response.text}"
        
        except requests.RequestException as e:
            return f"No se pudo conectar a Ollama: {e}"
    
    def save_analysis(self, anomalies, analysis):
        """Guarda el análisis en archivo JSON"""
        analysis_entry = {
            "timestamp": datetime.now().isoformat(),
            "anomalies_count": len(anomalies),
            "anomalies": anomalies,
            "llm_analysis": analysis
        }
        
        # Crear directorio si no existe
        Path("../logs").mkdir(parents=True, exist_ok=True)
        
        try:
            with open(self.analysis_file, 'a') as f:
                f.write(json.dumps(analysis_entry) + '\n')
            print(f"✓ Análisis guardado en: {self.analysis_file}")
        except Exception as e:
            print(f"❌ Error guardando análisis: {e}")
    
    def check_alert_conditions(self, anomalies, analysis):
        """Verifica si se debe generar alerta al administrador"""
        alert = {
            "should_alert": False,
            "reason": "",
            "severity": "LOW"
        }
        
        # Condiciones para alertas
        critical_keywords = [
            "certificado autofirmado",
            "certificado expirado",
            "intentos fallidos",
            "múltiples intentos",
            "patrón de ataque"
        ]
        
        analysis_lower = analysis.lower()
        
        # Contar anomalías
        anomaly_count = len(anomalies)
        
        # Verificar palabras clave críticas
        has_critical = any(keyword in analysis_lower for keyword in critical_keywords)
        
        # Lógica de alertas
        if anomaly_count >= 3 and has_critical:
            alert["should_alert"] = True
            alert["severity"] = "CRÍTICO"
            alert["reason"] = f"Múltiples anomalías críticas detectadas ({anomaly_count})"
        elif "crítico" in analysis_lower:
            alert["should_alert"] = True
            alert["severity"] = "CRÍTICO"
            alert["reason"] = "LLM detectó riesgo crítico"
        elif anomaly_count >= 2 and has_critical:
            alert["should_alert"] = True
            alert["severity"] = "ALTO"
            alert["reason"] = f"Múltiples anomalías de seguridad ({anomaly_count})"
        
        return alert
    
    def send_admin_notification(self, alert, analysis):
        """Envía notificación al administrador"""
        if not alert["should_alert"]:
            return
        
        notification = f"""
╔═══════════════════════════════════════════════════════════════╗
║         ⚠️  ALERTA DE SEGURIDAD mTLS - BANCO SANTANDER       ║
╚═══════════════════════════════════════════════════════════════╝

SEVERIDAD: {alert['severity']}
MOTIVO: {alert['reason']}
TIMESTAMP: {datetime.now().isoformat()}

ANÁLISIS DEL LLM:
{analysis}

ACCIÓN REQUERIDA:
1. Revisar inmediatamente los logs en: {self.log_file}
2. Verificar acceso de cuentas relacionadas
3. Considerar bloqueo de IPs sospechosas
4. Activar protocolos de incidente de seguridad

═══════════════════════════════════════════════════════════════
"""
        
        print(notification)
        
        # Guardar notificación en archivo
        try:
            with open("../logs/admin_alerts.log", 'a') as f:
                f.write(notification)
        except Exception as e:
            print(f"❌ Error guardando alerta: {e}")
    
    def run_analysis(self, minutes=60):
        """Ejecuta análisis completo"""
        print("\n" + "="*70)
        print("🔍 ANALIZADOR DE ANOMALÍAS mTLS CON LLM")
        print("="*70)
        
        # Verificar disponibilidad de Ollama
        print(f"\n🔌 Verificando conexión a Ollama ({self.ollama_url})...")
        if not self.check_ollama_availability():
            print(f"❌ Ollama no disponible en {self.ollama_url}")
            print("   Asegúrate de que Ollama está ejecutándose:")
            print("   $ ollama serve")
            print("   $ ollama pull llama2  # Si aún no tienes el modelo")
            return
        
        print(f"✓ Ollama conectado. Modelo: {self.model}")
        
        # Leer logs recientes
        print(f"\n📂 Leyendo logs recientes ({minutes} minutos)...")
        anomalies = self.read_recent_logs(minutes=minutes)
        print(f"✓ Se encontraron {len(anomalies)} anomalías")
        
        # Formatear para análisis
        anomalies_text = self.format_anomalies_for_analysis(anomalies)
        
        # Análisis con LLM
        print(f"\n🤖 Analizando con LLM ({self.model})...")
        analysis = self.analyze_with_llm(anomalies_text)
        
        print("\n" + "─"*70)
        print("ANÁLISIS DEL LLM:")
        print("─"*70)
        print(analysis)
        
        # Guardar análisis
        self.save_analysis(anomalies, analysis)
        
        # Verificar condiciones de alerta
        alert = self.check_alert_conditions(anomalies, analysis)
        
        # Enviar notificación si es necesario
        if alert["should_alert"]:
            self.send_admin_notification(alert, analysis)
        else:
            print("\n✓ No se requiere alerta al administrador")
        
        print("\n" + "="*70 + "\n")

def main():
    import argparse
    
    parser = argparse.ArgumentParser(
        description="Analizador de anomalías mTLS con LLM local"
    )
    parser.add_argument(
        "--ollama-url",
        default="http://localhost:11434",
        help="URL del servidor Ollama (default: http://localhost:11434)"
    )
    parser.add_argument(
        "--model",
        default="llama2",
        help="Modelo LLM a usar (default: llama2)"
    )
    parser.add_argument(
        "--minutes",
        type=int,
        default=60,
        help="Minutos atrás para analizar logs (default: 60)"
    )
    parser.add_argument(
        "--continuous",
        action="store_true",
        help="Ejecutar análisis continuamente cada 5 minutos"
    )
    
    args = parser.parse_args()
    
    analyzer = MTLSAnomalyAnalyzer(
        ollama_url=args.ollama_url,
        model=args.model
    )
    
    try:
        if args.continuous:
            print("🔄 Modo continuo activado. Presiona Ctrl+C para detener.")
            while True:
                analyzer.run_analysis(minutes=args.minutes)
                time.sleep(300)  # 5 minutos
        else:
            analyzer.run_analysis(minutes=args.minutes)
    except KeyboardInterrupt:
        print("\n\n⏹️  Análisis detenido por el usuario")
        sys.exit(0)

if __name__ == "__main__":
    main()
