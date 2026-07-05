
# 🏦 Banco Santander mTLS - Laboratorio de Seguridad

Simulación de servidor bancario con comunicaciones **mTLS (Mutual TLS)** bidireccionales, detección de anomalías en handshakes y análisis de seguridad con **LLM local (Ollama)**.

## 📋 Descripción del Proyecto

Este proyecto simula un **servidor bancario (Santander)** que se comunica de forma segura con múltiples **fintech (clientes)**:

- **Servidor**: Santander (Banco Virtual con Validación Bilateral)
- **Clientes Legítimos**:
  - Mercado Pago ✅
  - BBVA ✅
- **Cliente Atacante**: Sin certificado válido ❌

### Características Principales

✨ **mTLS (Mutual TLS)**
- Validación bidireccional de certificados
- Servidor y clientes se autentican mutuamente

🔐 **CA Raiz offline(BCRA)**
- Autoridad de Certificación de CAs Intermedias
- Certificados firmados y validables

🔐 **CA Intermedia (Banelco)**
- Autoridad de Certificación online
- Certificados clientes firmados y validables

📊 **Logging Estructurado**
- Registro de todos los handshakes TLS
- Formato JSONL para análisis
- Captura de anomalías en tiempo real

🤖 **Detección de Anomalías con LLM**
- Integración con Ollama (LLM local)
- Análisis automático de logs
- Identificación de patrones sospechosos

### Sistema
- **Go** 1.19+
- **Python** 3.8+
- **Ollama** (para análisis con LLM)
     Linux: curl -fsSL https://ollama.com/install.sh | sh
     Windows: https://ollama.com/download/windows
- **Ubuntu/Linux** (recomendado)

## 🚀 Inicio Rápido

### 1. Clonar el Repositorio

```bash
git clone https://github.com/JuanGallardo373/bancoSantander-mtls.git
cd bancoSantader-mtls
```
PKI de dos niveles con OCSP
```bash
git checkout main
```
PKI de un nivel sin OCSP
```bash
git checkout PKI-un-nivel
```

### 2. Generar Certificados

```bash
cd certs
bash CABancoCentral/generate-CARaiz.sh
bash CAIntermediaBANELCO/generate-CABanelco.sh
bash CABancoCentral/firmarCAIntermedia.sh

bash CAIntermediaBANELCO/createBundle.sh

bash cliente-mercadopago/bash generate-cert.sh
bash cliente-BBVA/generate-cert.sh
bash CAIntermediaBANELCO/firmarCertificados.sh
```

Esto generará:
- `bcra-raiz.key` y `bcra-raiz.crt` (Autoridad de Certificación Raiz)
- `banelco-inter.key` y `banelco-inter.crt` (Autoridad de Certificación Intermedia)
- `santander.key` y `santander.crt` (Servidor Santander)
- Certificados individuales para cada cliente
`
#OCSP
Descomentar function VerifyPeerCertificate en tls.Config en el archivo main.go
Comentar si no se utiliza OCSP
```bash
cd CAIntermediaBANELCO/
bash oscpKeyCSR.sh
bash signCertOCSP.sh
openssl ocsp -port 2560 -index index.txt -CA banelco-inter.crt -rkey ocsp.key -rsigner ocsp.crt
```
### 3. Iniciar el Servidor

```bash
go run servidor-banco/main.go
```

El servidor escuchará en `localhost:8443` con mTLS habilitado.

### 4. Analizar Logs con LLM (Descargar Ollama y llama3)

```bash
python analista-ia/llm-analyzer.py --ollama-url http://x.x.x.x:11434 #Modificar URL
```

### 5. Ejecutar Clientes (en otra terminal)

**Cliente Legítimo (Mercado Pago y BBVA):**
```bash
python3 cliente-mercadopago/cliente-mercadopago.py
python3 cliente-bancobbva/cliente-bbva.py
```

**Cliente Atacante:**
```bash
python3 cliente-atacante/cliente-autofirmado.py
python3 cliente-atacante/cliente-mismatch.py
```

📚 Referencias
Go TLS Documentation
Python Requests SSL
Ollama Documentation
mTLS Concepts
📄 Licencia
Este proyecto es parte de un laboratorio educativo de seguridad.

👤 Autor
Juan Gallardo - @JuanGallardo373
