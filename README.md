
# 🏦 Banco Santander mTLS - Laboratorio de Seguridad

Simulación de servidor bancario con comunicaciones **mTLS (Mutual TLS)** bidireccionales, detección de anomalías en handshakes y análisis de seguridad con **LLM local (Ollama)**.

## 📋 Descripción del Proyecto

Este proyecto implementa:

1. **Servidor Go** - Banco Santander con mTLS obligatorio
   - Endpoint `/transfer` para transferencias bancarias
   - Endpoint `/health` para verificar estado
   - Logging de anomalías en JSON
   - Validación de certificados cliente

2. **Clientes Python** - BBVA y Mercadopago
   - Comunicación mTLS con certificados cliente propios
   - Simulación de transferencias bancarias
   - Logging de transacciones

3. **Analizador de IA** - LLM local (Ollama)
   - Análisis automático de logs de anomalías
   - Detección de patrones de ataque
   - Notificaciones de seguridad al administrador
   - Clasificación de riesgos

## 🏗️ Estructura del Proyecto

```
bancoSantander-mtls/
├── servidor-banco/              # Servidor Go del banco
│   ├── main.go                 # Servidor mTLS principal
│   ├── certs/                  # Certificados del servidor
│   │   ├── ca-cert.pem        # Certificado de autoridad (compartido)
│   │   ├── servidor-cert.pem  # Certificado del servidor
│   │   └── servidor-key.pem   # Clave privada del servidor
│   └── go.mod
│
├── cliente-mercadopago/        # Cliente Python - Mercadopago
│   ├── cliente_mercadopago.py # Script del cliente
│   ├── certs/                 # Certificados de Mercadopago
│   │   ├── ca-cert.pem
│   │   ├── mercadopago-cert.pem
│   │   └── mercadopago-key.pem
│   └── requirements.txt
│
├── cliente-bancobbva/          # Cliente Python - BBVA
│   ├── cliente_bbva.py        # Script del cliente
│   ├── certs/                 # Certificados de BBVA
│   │   ├── ca-cert.pem
│   │   ├── bbva-cert.pem
│   │   └── bbva-key.pem
│   └── requirements.txt
│
├── analista-ia/               # Analizador con LLM
│   ├── llm_analyzer.py       # Analizador de anomalías
│   └── requirements.txt
│
├── logs/                      # Directorio de logs (generado)
│   ├── anomalies.jsonl       # Log de anomalías mTLS
│   ├── analysis.jsonl        # Resultados del análisis LLM
│   └── admin_alerts.log      # Alertas para administrador
│
└── README.md
```

## 🔐 Estructura de Certificados

Los certificados son autoauscriptos y están pre-generados:

```
CA (Certificate Authority)
├── ca-cert.pem    # Certificado de la CA (validará todos los clientes)
└── ca-key.pem     # Clave privada de la CA

Servidor Santander
├── servidor-cert.pem  # Certificado del servidor (firmado por CA)
└── servidor-key.pem   # Clave privada del servidor

Cliente BBVA
├── bbva-cert.pem   # Certificado del cliente (firmado por CA)
└── bbva-key.pem    # Clave privada del cliente

Cliente Mercadopago
├── mercadopago-cert.pem  # Certificado del cliente (firmado por CA)
└── mercadopago-key.pem   # Clave privada del cliente
```

## 🚀 Requisitos Previos

### Sistema
- **Go** 1.19+ 
- **Python** 3.8+
- **Ollama** (para análisis con LLM)

### Instalación

```bash
# Clonar repositorio
git clone https://github.com/JuanGallardo373/bancoSantander-mtls.git
cd bancoSantander-mtls

# Instalar dependencias Python (en cada directorio cliente y analista)
pip install -r cliente-mercadopago/requirements.txt
pip install -r cliente-bancobbva/requirements.txt
pip install -r analista-ia/requirements.txt

# Instalar y descargar modelo Ollama
ollama pull llama2
```

## 🎯 Cómo Ejecutar

### 1️⃣ Iniciar el Servidor Santander

```bash
cd servidor-banco
go run main.go
```

**Salida esperada:**
```
🚀 Servidor Banco Santander iniciado en https://localhost:8443
📋 Esperando solicitudes con mTLS...
📊 Logs de anomalías guardados en: ../logs/anomalies.jsonl
```

### 2️⃣ Ejecutar Cliente Mercadopago

En otra terminal:

```bash
cd cliente-mercadopago
python cliente_mercadopago.py
```

**Salida esperada:**
```
============================================================
🎯 Cliente Mercadopago - mTLS Bank Transfer
============================================================
Conectando a: https://localhost:8443
Certificados: ./certs/mercadopago-cert.pem, ./certs/mercadopago-key.pem

🔐 Configurando mTLS...
✓ Verificando conectividad con el servidor...
✓ Servidor activo: {...}

============================================================
📤 Enviando transferencias...
============================================================
...
```

### 3️⃣ Ejecutar Cliente BBVA

En otra terminal:

```bash
cd cliente-bancobbva
python cliente_bbva.py
```

### 4️⃣ Analizar Anomalías con LLM

En otra terminal (con Ollama ejecutándose):

```bash
# Asegurar que Ollama está corriendo
ollama serve

# En otra terminal
cd analista-ia
python llm_analyzer.py
```

**Opciones avanzadas:**

```bash
# Analizar últimos 30 minutos
python llm_analyzer.py --minutes 30

# Usar modelo diferente
python llm_analyzer.py --model mistral

# Modo continuo (analiza cada 5 minutos)
python llm_analyzer.py --continuous
```

## 📊 Flujo de Datos

```
Cliente BBVA                  Cliente Mercadopago
    │                                │
    │ (mTLS + cert)                  │ (mTLS + cert)
    │                                │
    └────────────────┬───────────────┘
                     │
                     ▼
            🏦 Servidor Santander
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
    Transfer    Health      Validate Cert
    Process     Check       (Log Anomalies)
        │                        │
        └────────────┬───────────┘
                     │
                     ▼
            📂 logs/anomalies.jsonl
                     │
                     ▼
        🤖 Ollama LLM Analyzer
                     │
        ┌────────────┴────────────┐
        ▼                         ▼
    Analysis              Admin Alerts
    logs/analysis.jsonl   logs/admin_alerts.log
```

## 🔍 Tipos de Anomalías Detectadas

El servidor detecta y registra automáticamente:

### 1. **Certificado Expirado**
- Cliente intenta conectar con certificado que pasó su fecha de validez
- **Severity**: ALTO
- **Action**: Rechazar conexión, loguear evento

### 2. **Certificado Autofirmado**
- Cliente presenta certificado que no fue firmado por la CA
- **Severity**: CRÍTICO
- **Action**: Rechazar conexión inmediatamente

### 3. **Certificado Inválido**
- Cliente sin certificado o certificado corrupto
- **Severity**: CRÍTICO
- **Action**: Rechazar handshake

### 4. **Patrones de Ataque**
- Múltiples intentos fallidos desde misma IP
- Certificados falsos secuenciales
- **Detection**: LLM analiza y detecta patrones

## 📈 Ejemplo de Log de Anomalía

```json
{
  "timestamp": "2024-05-26T14:35:22.123456",
  "event_type": "SUSPICIOUS_CERTIFICATE",
  "client_ip": "192.168.1.100",
  "client_name": "attacker.com",
  "error_message": "Certificado autofirmado detectado; Certificado vencido detectado",
  "is_expired": true,
  "is_self_signed": true
}
```

## 🤖 Ejemplo de Análisis LLM

```
CLASIFICACIÓN DE RIESGO: CRÍTICO

ANÁLISIS POR ANOMALÍA:
1. Cliente autofirmado desde 192.168.1.100
   - Intento obvio de suplantación
   - Patrón típico de ataque MITM
   - NO AUTORIZAR bajo ninguna circunstancia

PATRONES DETECTADOS:
- 3 intentos fallidos en 2 minutos desde misma IP
- Certificados con nombres de dominio inconsistentes
- Uso de protocolos de cifrado débiles

RECOMENDACIONES:
✓ BLOQUEAR IP inmediatamente: 192.168.1.100
✓ Alertar al equipo de seguridad
✓ Revisar logs de acceso de últimas 24h
✓ Considerar honeypot para capturar más intel

NOTIFICAR AL ADMINISTRADOR: SÍ
```

## 🛡️ Características de Seguridad

✅ **mTLS Obligatorio**
- Requiere certificado cliente válido
- Valida fecha de expiración
- Verifica firma de CA

✅ **Logging Detallado**
- Todos los errores de handshake registrados
- Formato JSON para parsing automático
- Timestamps precisos

✅ **Análisis Inteligente**
- LLM detecta patrones de ataque
- Clasificación automática de riesgos
- Alertas contextualizadas

✅ **Notificaciones**
- Alertas para administrador en tiempo real
- Recomendaciones de acción
- Severidad clasificada

## 🔧 Troubleshooting

### Error: "Error cargando CA certificate"
```
Solución: Verificar que ca-cert.pem existe en servidor-banco/certs/
```

### Error: "SSL: CERTIFICATE_VERIFY_FAILED"
```
Solución: Asegurar que el certificado cliente es válido y está firmado por la CA
```

### Error: "No se puede conectar a Ollama"
```
Solución: 
1. Instalar Ollama: https://ollama.ai
2. Ejecutar: ollama serve
3. Descargar modelo: ollama pull llama2
```

### Error: "Connection refused" en cliente
```
Solución: Asegurar que el servidor Go está ejecutándose en puerto 8443
```

## 📝 Notas de Desarrollo

### Agregar nuevo cliente
1. Generar certificado con autoridad CA
2. Crear script Python en `cliente-{banco}/`
3. Usar mismo patrón de MTLSAdapter
4. Actualizar documentación

### Modificar endpoint
1. Editar función handler en `main.go`
2. Agregar validaciones necesarias
3. Loguear anomalías detectadas
4. Recompilar: `go run main.go`

### Personalizar análisis LLM
1. Modificar `prompt` en `analyze_with_llm()`
2. Ajustar temperatura (0-1) para variabilidad
3. Cambiar modelo en parámetro `--model`

## 📚 Referencias

- [Go TLS Documentation](https://golang.org/pkg/crypto/tls/)
- [Python Requests SSL](https://docs.python-requests.org/en/latest/user/advanced/#ssl-cert-verification)
- [Ollama Documentation](https://github.com/jmorganca/ollama)
- [mTLS Concepts](https://www.cloudflare.com/learning/access-management/what-is-mtls/)

## 📄 Licencia

Este proyecto es parte de un laboratorio educativo de seguridad.

## 👤 Autor

Juan Gallardo - [@JuanGallardo373](https://github.com/JuanGallardo373)

---

**⚠️ Nota Importante**: Este proyecto es para fines educativos. Los certificados son auto-autofirmados y no deben usarse en producción.
