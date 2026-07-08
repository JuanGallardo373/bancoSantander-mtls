# 🏦 Laboratorio con infraestructura PKI jerárquica, validación OCSP y análisis local de anomalías para entornos financieros mTLS

Simulación de servidor bancario con comunicaciones **mTLS (Mutual TLS)** con PKI jerarquica y validación OCSP, detección de anomalías en handshakes y análisis de seguridad con **LLM local (Ollama)**.

## 🛠️ Descripción del Pipeline del Sistema

El proyecto implementa un pipeline de datos y seguridad dividido en cuatro etapas secuenciales y desacopladas:
1. **Capa Criptográfica de Transporte (mTLS):** Autenticación bidireccional basada de certificados X.509 que valida la cadena de confianza de 2 niveles (CA Raíz y CA Intermedia) exigiendo la extensión `clientAuth`.
2. **Validación Dinámica (OCSP):** Intercepción del handshake para interrogar síncronamente al respondedor OpenSSL sobre la vigencia operativa del certificado del cliente en tiempo real.
3. **Ingesta de Observabilidad:** Almacenamiento estructurado de errores de negociación, *resets* de sockets y anomalías criptográficas directamente en registros binarios a texto plano dentro de `logs/anomalies.jsonl`.
4. **Auditoría Forense Avanzada (IA):** Pipeline diferido donde los modelos compactos locales consumen los registros logs para interpretar sintácticamente las anomalías y generar reportes ejecutivos automatizados de ciberseguridad.

## Instalación y requisitos de hardware
- **Go** 1.19+
sudo apt install golang-go

- **Python** 3.8+
sudo apt install -y python3 python3-pip python3-dev python3-venv libssl-dev libffi-dev && pip3 install --upgrade pip

- **Ollama** (para análisis con LLM)
Linux: curl -fsSL https://ollama.com/install.sh | sh
Windows: https://ollama.com/download/windows

- **OCSP** (para revocacion de certs)
go get golang.org/x/crypto/ocsp

- **Ubuntu/Linux** (recomendado)

## Modelos Compactos Evaluados

Para el componente de monitoreo y auditoría forense se priorizó el uso de **Modelos de Lenguaje de Tamaño Reducido o Compactos (SLMs)**. Estos modelos permiten una ejecución 100% local y *offline* en la infraestructura del servidor, garantizando la privacidad del secreto bancario al no derivar registros de transacciones a APIs de terceros en la nube.

Se evaluaron y contrastaron las siguientes dos arquitecturas utilizando la plataforma **Ollama** con cuantización nativa a 4 bits (Formatos GGUF - `q4_K_M`):

| Modelo Evaluado | Tamaño (Parámetros) | Arquitectura Base | Propósito en el Proyecto |
|---|---|---|---|
| **Llama 3** | 8B | Meta-Llama-3 | Modelo principal para la interpretación semántica y generación de reportes ejecutivos automatizados en español. |
| **Llama 2** | 7B | Meta-Llama-2 | Modelo de control comparativo utilizado para evaluar la evolución en la precisión de inferencia sobre logs estructurados. |

## Comandos de Ejecución
### 1. Clonar el Repositorio

```bash
git clone https://github.com/JuanGallardo373/bancoSantander-mtls.git
cd bancoSantander-mtls
```
Branch con PKI de dos niveles con OCSP
```bash
git checkout main
```
Branch con PKI de un nivel sin OCSP
```bash
git checkout PKI-un-nivel
```

### 2.1 Generar Certificados (main -> PKI de dos niveles)

```bash
cd CABancoCentral/
bash generate-CARaiz.sh

cd CAIntermediaBANELCO/
bash generate-CABanelco.sh

cd CABancoCentral/
bash firmarCAIntermedia.sh

cd CAIntermediaBANELCO/
bash createBundle.sh

cd servidor-banco/
bash generate-cert.sh

cd cliente-mercadopago/
bash generate-cert.sh

cd cliente-BBVA/
bash generate-cert.sh

cd CAIntermediaBANELCO/
bash firmarCertificados.sh
```

Esto generará:
- `bcra-raiz.key` y `bcra-raiz.crt` (Autoridad de Certificación Raiz)
- `banelco-inter.key` y `banelco-inter.crt` (Autoridad de Certificación Intermedia)
- `santander.key` y `santander.crt` (Servidor Santander)
- Certificados individuales para cada cliente

```bash
cd cliente-atacante/
bash cert-atacante.sh
```
Esto generará la clave privada (atacante.key) y el certificado autofirmado del atacante (atacante.crt)

-**OCSP**
- Descomentar function VerifyPeerCertificate en tls.Config en el archivo main.go
- Comentar si no se utiliza OCSP
```bash
cd CAIntermediaBANELCO/
bash oscpKeyCSR.sh
bash signCertOCSP.sh
bash iniciarServidorOCSP.sh
```

### 2.2 Generar Certificados (PKI-un-nivel)

```bash
cd CABancoCentral/
bash generate-CARaiz.sh

cd servidor-banco/
bash generate-cert.sh

cd cliente-mercadopago/
bash generate-cert.sh

cd cliente-BBVA/
bash generate-cert.sh

cd CABancoCentral/
bash firmarCertificados.sh
```

Esto generará:
- `cakey.pem` y `cacert.pem` (Autoridad de Certificación)
- `santander.key` y `santander.crt` (Servidor Santander)
- Certificados individuales para cada cliente

```bash
cd cliente-atacante/
bash cert-atacante.sh
```
Esto generará la clave privada (atacante.key) y el certificado autofirmado del atacante (atacante.crt)

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

## Descripción de los datasets
**logs/**
- anomalies.jsonl: Registro de anomalías y errores en los handshake TLS 1.3 detectados por el servidor con información acerca del error, IP del cliente, timestamp, etc.
- analysis.jsonl: Análisis de los logs en anomalies.jsonl generado por el LLM con detección de patrones, tipo de ataque, clasificación de riesgo y recomendaciones de seguridad para el equipo de ciberseguridad.
- admin_alerts.log: Alertas para el administrador generadas por el LLM

**data/**
- Resultados obtenidos en la latencia de los handshakes durante las 10 corridas por cada infraestructura de la PKI y cálculos de Media, Desvio Estándar, Intervalo de confianza en cada escenario. 

## Metodología

El enfoque metodológico seleccionado para este trabajo es el de Simulación de
Entornos Distribuidos mediante Implementación de Software. Este diseño
experimental permite reproducir con exactitud matemática y criptográfica el
comportamiento de una red interbancaria real, abstrayendo la complejidad del
hardware físico a través de un aislamiento lógico de procesos en el sistema operativo
de una Máquina Virtual corriendo bajo Linux.
1. Fidelidad del Stack de Protocolos: Los scripts de software desarrollados
interactúan de manera directa con las llamadas al sistema (syscalls) del kernel
de Linux para la creación de sockets TCP/IP reales. Al utilizar interfaces de red
locales (loopback / 127.0.0.1), las librerías de Go y Python ejecutan los
algoritmos criptográficos (ECDHE, firmas X.509) de manera idéntica a como
ocurriría a través de un cable de red físico entre servidores remotos. Esto
permite capturar y analizar el tráfico real mediante herramientas de inspección
de paquetes.
2. Aislamiento y Control de Variables: La simulación permite estructurar un
entorno controlado mediante directorios independientes que emulan el
aislamiento perimetral. Al compartimentar las claves privadas de las
Autoridades Certificadoras, el Servidor Bancario y los Clientes en directorios
aislados, se garantiza el cumplimiento estricto del principio de menor privilegio,
asegurando que ninguna entidad acceda a la memoria física o los archivos de
configuración de las otras.
3. Seguridad y Repetibilidad del Escenario de Ataque: Ejecutar las pruebas
con un Cliente Atacante provisto de un certificado falsificado permite analizar el
comportamiento del backend ante fallos críticos de seguridad sin poner en
riesgo una infraestructura de producción real. La naturaleza del software
desarrollado garantiza la repetibilidad exacta del experimento, facilitando la
recolección y consistencia de los registros de log necesarios para evaluar la
precisión del analizador basado en inteligencia artificial.

## Resultados y Limitaciones

Los resultados en cuanto al overhead en los handshakes que se obtuvieron luego de las pruebas con los distintos escenarios de la PKI fueron los siguientes:

| Arquitectura de la PKI | Media | Desvío Estándar | Intervalo de Confianza 95% [Lím. Inf - Lím. Sup] |
|---|---|---|---|
| Escenario A: PKI de un nivel sin OCSP | 5,22 ms | 0,55 ms | [4,83 ms; 5,61 ms] |
| Escenario B: PKI de dos niveles sin OCSP | 7,47 ms | 1,05 ms | [6,72 ms; 8,22 ms] |
| Escenario C: PKI de dos niveles con OCSP | 14,84 ms | 0,75 ms | [13,49 ms; 16,20 ms] |

- El script que se utilizo para estas mediciones fue medir_latencia.sh en la rama **main** para medir el Escenario B y C, y en la rama **PKI-un-nivel** para el Escenario A.
```bash
chmod +x medir_latencia.sh
./medir_latencia
```
- Por cada escenario se hicieron 10 corridas para mayor precisión en los datos
- Estos datos se pueden visualizar en el archivo data/pki_comparison_results.md 

La PKI de dos niveles con OCSP introduce un overhead de 9,62 ms en el handshake TLS 1.3 respecto a la PKI de un nivel, manteniendo compatibilidad completa con los clientes existentes y cerrando el vector de reutilización de certificados comprometidos

En cuanto a la latencia en los handshakes TLS con los distintos clientes (legitimos y atacantes) y el tiempo de respuesta del LLM se obtuvieron los siguientes resultados:
| Escenario | Latencia del Handshake TLS | Tiempo de inferencia del LLM |
|---|---|---|
| Cliente Legítimo (Mercado Pago / BBVA) con cadena válida | ≈9 ms | 0.00 s (Ignorado por diseño) |
| Atacante con Certificado Autofirmado (Suplantación) | ≈35 ms | ≈ 24 s | 
| Atacante con Robo Parcial (Mismatch) | ≈7 ms | ≈23 s |



### Limitaciones identificadas
A partir de las pruebas de estrés intensivo y el análisis de comportamiento bajo
concurrencia simétrica realizados sobre la plataforma, se identificaron dos limitaciones
técnico-arquitectónicas críticas en el diseño actual de la solución:
1. Bloqueo Síncrono y Monohilo del Respondedor OCSP Nativo:
La validación del estado de revocación de los certificados de los clientes se ejecuta de
forma síncrona dentro del ciclo de vida del handshake mutuo de la capa de transporte
(mTLS) mediante el callback VerifyPeerCertificate del servidor en Go. Si bien el
backend de la API aprovecha el modelo de concurrencia basado en goroutines nativas
del lenguaje, el servidor OCSP provisto por la suite de OpenSSL opera sobre un único
hilo de ejecución y realiza accesos secuenciales a un archivo estructurado en disco
(index.txt). Bajo escenarios de alta disponibilidad y ráfagas masivas de peticiones
concurrentes, el respondedor OCSP se transforma en un cuello de botella crítico de la
infraestructura distribuida. Esta asimetría operativa degrada el rendimiento global del
sistema, induciendo la saturación de los buffers de red del kernel, el agotamiento de
los descriptores de archivos (file descriptors) y la consecuente aparición de fallos por
expiración de tiempo de espera (context deadline exceeded), abortando las
conexiones legítimas antes de completar la negociación criptográfica.
2. Incompatibilidad del Procesamiento en Tiempo Real de Auditoría de IA:
El componente de monitoreo y auditoría basado en modelos de lenguaje extenso
(LLM) presenta una limitación operativa inherente respecto a los tiempos de respuesta
exigidos por la arquitectura interbancaria. Aunque la ingesta y el análisis semántico de
los registros de anomalías estructurados en el archivo anomalies.jsonl son altamente
eficaces para clasificar fallos de negociación criptográfica complejos, la latencia de
inferencia del modelo impide su utilización en caliente (en línea con la transferencia).
El costo computacional en términos de tiempo por token generado restringe la
intervención del LLM exclusivamente al plano operativo offline o en diferido
(procesamiento por lotes). Por ende, el sistema carece de capacidad para mitigar
ataques de denegación de servicio (DoS) por suplantación criptográfica en el instante
preciso en que ocurren, limitando su rol a un esquema puramente analítico, reactivo y
de auditoría forense posterior.

## Autores y Contexto Académico
Juan Ignacio Gallardo : autor principal y desarrollador.
Benjamín Chuquimango: coautor y director académico
Universidad Nacional de General Sarmiento, 2026.
Trabajo Final Individual de Sistemas Operativos y Redes 2, primer semestre de 2026.

## Estado y licencia
Licencia: pendiente de definición antes de una eventual publicación pública.
Citación: pendiente de definición después de la revisión académi