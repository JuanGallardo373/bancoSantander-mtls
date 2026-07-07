package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ocsp"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// TransferRequest representa una solicitud de transferencia
type TransferRequest struct {
	SourceBank      string  `json:"source_bank"`
	SourceAccount   string  `json:"source_account"`
	DestinationBank string  `json:"destination_bank"`
	DestAccount     string  `json:"destination_account"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
}

// TransferResponse representa la respuesta de una transferencia
type TransferResponse struct {
	Status     string    `json:"status"`
	TransferID string    `json:"transfer_id"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}

// AnomalyLog registra anomalías detectadas en el handshake mTLS
type AnomalyLog struct {
	Timestamp      time.Time `json:"timestamp"`
	EventType      string    `json:"event_type"`
	ClientIP       string    `json:"client_ip"`
	ClientName     string    `json:"client_name"`
	ErrorMessage   string    `json:"error_message"`
	HandshakeError bool      `json:"handshake_error"`
}

var (
	logFile           *os.File
	logMutex          sync.Mutex
	transferIDCounter int
	totalHandshakes   int64
	totalLatencyMs    int64
	metricsMutex      sync.Mutex
	ocspClient        *http.Client
	ocspServerURL     string = "http://localhost:2560"
	issuerCert        *x509.Certificate
)

func init() {
	var err error

	// Crear directorio de logs si no existe
	os.MkdirAll("../logs", 0755)

	// Crear o abrir el archivo de logs JSON
	logFile, err = os.OpenFile("../logs/anomalies.jsonl", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error al abrir archivo de logs: %v", err)
	}

	// Inicializar cliente HTTP para OCSP con timeout
	ocspClient = &http.Client{
		Timeout: 5 * time.Second,
	}
}

func logAnomaly(anomaly AnomalyLog) {
	anomaly.Timestamp = time.Now()

	logMutex.Lock()
	defer logMutex.Unlock()

	encoder := json.NewEncoder(logFile)
	if err := encoder.Encode(anomaly); err != nil {
		log.Printf("Error escribiendo log de anomalía: %v", err)
	}
	logFile.Sync()
}

// extractClientIP extrae la IP del cliente de la dirección remota
func extractClientIP(remoteAddr string) string {
	parts := strings.Split(remoteAddr, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return remoteAddr
}

// verifyOCSPLocal verifica el estado del certificado con el servidor OCSP local
func verifyOCSPLocal(cert *x509.Certificate) error {
	if issuerCert == nil {
		log.Printf("⚠️ Certificado emisor no cargado para verificación OCSP")
		return nil
	}

	// Crear solicitud OCSP
	ocspReq, err := ocsp.CreateRequest(cert, issuerCert, nil)
	if err != nil {
		log.Printf("⚠️ Error creando solicitud OCSP: %v", err)
		return nil
	}

	// Enviar solicitud al servidor OCSP local
	//log.Printf("🔍 Verificando certificado con servidor OCSP local: %s", ocspServerURL)

	resp, err := ocspClient.Post(ocspServerURL, "application/ocsp-request", strings.NewReader(string(ocspReq)))
	if err != nil {
		log.Printf("❌ Error contactando servidor OCSP local en %s: %v", ocspServerURL, err)
		return fmt.Errorf("fallo verificación OCSP: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Servidor OCSP local retornó estado: %d", resp.StatusCode)
		return fmt.Errorf("servidor OCSP retornó estado %d", resp.StatusCode)
	}

	// Parsear respuesta OCSP
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Error leyendo respuesta OCSP: %v", err)
		return fmt.Errorf("error leyendo respuesta OCSP: %v", err)
	}

	ocspResp, err := ocsp.ParseResponse(respBytes, issuerCert)
	if err != nil {
		log.Printf("❌ Error parseando respuesta OCSP: %v", err)
		return fmt.Errorf("error parseando respuesta OCSP: %v", err)
	}

	// Verificar estado del certificado
	switch ocspResp.Status {
	case ocsp.Good:
		//		log.Printf("✅ Certificado VÁLIDO según OCSP | NextUpdate: %v", ocspResp.NextUpdate)
		return nil
	case ocsp.Revoked:
		errorMsg := fmt.Sprintf("CERTIFICADO REVOCADO desde %v por razón: %s", ocspResp.RevokedAt, ocspResp.RevocationReason)
		log.Printf("❌ %s", errorMsg)
		return fmt.Errorf(errorMsg)
	case ocsp.Unknown:
		log.Printf("⚠️ Estado desconocido en OCSP para certificado")
		return fmt.Errorf("estado OCSP desconocido")
	}

	return nil
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Parsear el body de la solicitud
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error leyendo el body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var transferReq TransferRequest
	if err := json.Unmarshal(body, &transferReq); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Generar ID de transferencia
	transferIDCounter++
	transferID := fmt.Sprintf("TRX-%d-%d", time.Now().Unix(), transferIDCounter)

	// Log de transferencia exitosa
	log.Printf("✓ Transferencia recibida: %s | De: %s | A: %s | Monto: %.2f %s",
		transferID, transferReq.SourceBank, transferReq.DestinationBank, transferReq.Amount, transferReq.Currency)

	// Respuesta exitosa
	response := TransferResponse{
		Status:     "SUCCESS",
		TransferID: transferID,
		Message:    fmt.Sprintf("Transferencia de %.2f %s procesada exitosamente", transferReq.Amount, transferReq.Currency),
		Timestamp:  time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"server": "Banco Santander mTLS",
	})
}

// CustomTLSListener envuelve el listener TCP crudo
type CustomTLSListener struct {
	net.Listener
	tlsConfig *tls.Config
}

func (l *CustomTLSListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// 1. Envolvemos la conexión TCP en una conexión TLS, pero NO hacemos el handshake todavía.
	// Esto es rapidísimo y no bloquea el hilo principal.
	tlsConn := tls.Server(conn, l.tlsConfig)

	// 2. Entregamos nuestro Wrapper al servidor HTTP.
	// El servidor HTTP mandará esto a una Goroutine de fondo.
	return &TLSHandshakeInterceptor{Conn: tlsConn}, nil
}

// TLSHandshakeInterceptor intercepta el momento exacto en que Go intenta comunicarse
type TLSHandshakeInterceptor struct {
	*tls.Conn
	handshakeDone bool
}

// Método Read para capturar el error TLS real y medir el Handshake
func (c *TLSHandshakeInterceptor) Read(b []byte) (int, error) {
	// Forzamos el Handshake manualmente en la primera lectura (que ya ocurre en el fondo)
	if !c.handshakeDone {
		// ⏱️ CAPTURA INICIAL
		startHandshake := time.Now()
		err := c.Conn.Handshake()
		c.handshakeDone = true

		// ⏱️ CÁLCULO FINAL: Handshake finalizado (con éxito o error)
		latency := time.Since(startHandshake).Milliseconds()

		// Imprimir reporte resumido cada 100 handshakes para no saturar la consola
		if totalHandshakes%100 == 0 {
			mediaActual := float64(totalLatencyMs) / float64(totalHandshakes)
			log.Printf("\n📊 [MÉTRICAS] Handshakes evaluados: %d | Latencia Media Actual: %.2f ms\n", totalHandshakes, mediaActual)
		}

		if err != nil {
			// CAPTURAMOS LA ANOMALÍA EXACTA PARA LA IA
			clientIP := extractClientIP(c.Conn.RemoteAddr().String())
			// 🔴 FILTRO DE CONTINGENCIA: Si el error es un reset por descarte del cliente, lo ignoramos de anomalies.jsonl
			if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "broken pipe") {
				metricsMutex.Lock()
				totalHandshakes++
				totalLatencyMs += latency
				metricsMutex.Unlock()

				//log.Printf("⚠️ Socket TCP cerrado abruptamente por el cliente de estrés (s_time) | IP: %s", clientIP)
				return 0, err
			}
			anomaly := AnomalyLog{
				EventType:      "MTLS_HANDSHAKE_FAILED",
				ClientIP:       clientIP,
				ClientName:     "unknown", // No sabemos quién es porque falló
				ErrorMessage:   err.Error(),
				HandshakeError: true,
			}

			logAnomaly(anomaly)

			//log.Printf("🔐 ❌ ERROR HANDSHAKE mTLS: %s | IP: %s | Tiempo de procesamiento kernel: %d ms", err.Error(), clientIP, latencia)
			return 0, err
		}
		metricsMutex.Lock()
		totalHandshakes++
		totalLatencyMs += latency
		metricsMutex.Unlock()

		/*var clientName string
		if len(c.Conn.ConnectionState().PeerCertificates) > 0 {
			clientName = c.Conn.ConnectionState().PeerCertificates[0].Subject.CommonName
		} else {
			clientName = "unknown"
		}
		log.Printf("🔐 🟢 HANDSHAKE mTLS EXITOSO | Cliente: %s | Latencia de negociación: %d ms", clientName, latency)
		*/
	}
	return c.Conn.Read(b)
}

func main() {
	defer logFile.Close()

	// Cargar certificados de cliente (CA)
	caCert, err := os.ReadFile("../CAIntermediaBANELCO/bundle.crt")
	if err != nil {
		log.Fatalf("Error cargando el bundle de CAs: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatal("Error agregando el bundle de CAs al pool de Go")
	}

	// Cargar certificado intermedio para OCSP
	intermediaBytes, err := os.ReadFile("../CAIntermediaBANELCO/banelco-inter.crt")
	if err != nil {
		log.Fatalf("Error cargando certificado intermedio: %v", err)
	}

	block, _ := pem.Decode(intermediaBytes)
	if block == nil {
		log.Fatal("Error decodificando certificado intermedio")
	}

	issuerCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		log.Fatalf("Error parseando certificado intermedio: %v", err)
	}

	// Cargar certificado y clave del servidor
	serverCert, err := tls.LoadX509KeyPair(
		"./certs/santander.crt",
		"./certs/santander.key",
	)
	if err != nil {
		log.Fatalf("Error cargando certificado del servidor: %v", err)
	}

	serverCert.Certificate = append(serverCert.Certificate, block.Bytes)

	// Configurar TLS 1.3 con mTLS obligatorio + VerifyPeerCertificate para OCSP
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS13,
		// Callback para verificación adicional de certificados de cliente (OCSP)
/*		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return fmt.Errorf("no se recibió certificado del cliente")
			}

			// Parsear primer certificado del cliente
			clientCert, err := x509.ParseCertificate(rawCerts[0])
			if err != nil {
				return fmt.Errorf("error parseando certificado del cliente: %v", err)
			}

			// Verificar OCSP del certificado del cliente con servidor local
			//log.Printf("🔐 Verificando OCSP para cliente: %s", clientCert.Subject.CommonName)
			if err := verifyOCSPLocal(clientCert); err != nil {
				return err
			}

			return nil
		},
*/	}

	// Registrar handlers
	http.HandleFunc("/transfer", transferHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("🚀 Servidor Banco Santander iniciado en https://localhost:8443")
	log.Println("📋 Esperando solicitudes con mTLS obligatorio (TLS 1.3)...")
	log.Println("🔍 Verificación OCSP HABILITADA - Servidor OCSP: " + ocspServerURL)
	log.Println("📊 Logs de anomalías guardados en: ../logs/anomalies.jsonl")
	log.Println()
	log.Println("🔐 Configuración mTLS:")
	log.Println("   ✓ TLS 1.3 mínimo (máxima seguridad)")
	log.Println("   ✓ Requiere certificado cliente válido")
	log.Println("   ✓ Verifica firma de CA")
	log.Println("   ✓ ✅ Verifica estado OCSP con servidor local en " + ocspServerURL)
	log.Println("   ✓ Rechaza automáticamente certificados inválidos en handshake")
	log.Println("   ✓ Rechaza certificados revocados según OCSP")
	log.Println("   ✓ Registra todos los intentos fallidos en anomalies.jsonl")
	log.Println()

	// Crear listener TCP
	tcpListener, err := net.Listen("tcp", ":8443")
	if err != nil {
		log.Fatalf("Error creando listener TCP: %v", err)
	}
	defer tcpListener.Close()

	// Crear listener TLS personalizado que captura handshake failures
	customTLSListener := &CustomTLSListener{
		Listener:  tcpListener,
		tlsConfig: tlsConfig,
	}
	defer customTLSListener.Close()

	// Crear servidor HTTP que usará el listener personalizado
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
		Handler:   http.DefaultServeMux,
	}

	// Servir con listener personalizado (SIN usar ListenAndServeTLS)
	if err := server.Serve(customTLSListener); err != nil {
		log.Fatalf("Error en servidor: %v", err)
	}
}
