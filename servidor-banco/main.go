package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
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
	Status      string    `json:"status"`
	TransferID  string    `json:"transfer_id"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
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

func transferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtener nombre del cliente del certificado validado
	var clientName string
	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		clientName = r.TLS.PeerCertificates[0].Subject.CommonName
	} else {
		clientName = "unknown"
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
	log.Printf("✓ Transferencia recibida: %s | De: %s | A: %s | Monto: %.2f %s | Cliente: %s",
		transferID, transferReq.SourceBank, transferReq.DestinationBank, transferReq.Amount, transferReq.Currency, clientName)

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

// TLSListenerWrapper intercepta errores de handshake TLS
type TLSListenerWrapper struct {
	net.Listener
}

func (l *TLSListenerWrapper) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &TLSConnWrapper{Conn: conn}, nil
}

// TLSConnWrapper intercepta el handshake TLS y registra errores
type TLSConnWrapper struct {
	net.Conn
	wrapped bool
}

func (c *TLSConnWrapper) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)

	// Si hay error en el handshake TLS, registrarlo
	if err != nil && !c.wrapped && strings.Contains(err.Error(), "tls:") {
		c.wrapped = true
		clientIP := extractClientIP(c.Conn.RemoteAddr().String())

		anomaly := AnomalyLog{
			EventType:      "MTLS_HANDSHAKE_ERROR",
			ClientIP:       clientIP,
			ClientName:     "unknown",
			ErrorMessage:   err.Error(),
			HandshakeError: true,
		}

		logAnomaly(anomaly)
		log.Printf("🔐 ERROR HANDSHAKE mTLS: %s | IP: %s", err.Error(), clientIP)
	}

	return n, err
}

func main() {
	defer logFile.Close()

	// Cargar certificados de cliente (CA)
	caCert, err := os.ReadFile("./certs/ca-cert.pem")
	if err != nil {
		log.Fatalf("Error cargando CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatal("Error agregando CA certificate al pool")
	}

	// Cargar certificado y clave del servidor
	serverCert, err := tls.LoadX509KeyPair(
		"./certs/servidor-cert.pem",
		"./certs/servidor-key.pem",
	)
	if err != nil {
		log.Fatalf("Error cargando certificado del servidor: %v", err)
	}

	// Configurar TLS 1.3 con mTLS obligatorio
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert, // Rechaza si certificado es inválido
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS13, // TLS 1.3 mínimo
		// CipherSuites no es necesario en TLS 1.3 (Go usa automáticamente los 5 suites estándar)
	}

	// Registrar handlers
	http.HandleFunc("/transfer", transferHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("🚀 Servidor Banco Santander iniciado en https://localhost:8443")
	log.Println("📋 Esperando solicitudes con mTLS obligatorio (TLS 1.3)...")
	log.Println("📊 Logs de anomalías guardados en: ../logs/anomalies.jsonl")
	log.Println()
	log.Println("🔐 Configuración mTLS:")
	log.Println("   ✓ TLS 1.3 mínimo (máxima seguridad)")
	log.Println("   ✓ Requiere certificado cliente válido")
	log.Println("   ✓ Verifica firma de CA")
	log.Println("   ✓ Rechaza automáticamente certificados inválidos en handshake")
	log.Println("   ✓ Registra todos los intentos fallidos en anomalies.jsonl")
	log.Println()

	// Crear listener TCP
	tcpListener, err := net.Listen("tcp", ":8443")
	if err != nil {
		log.Fatalf("Error creando listener TCP: %v", err)
	}
	defer tcpListener.Close()

	// Envolver con TLS
	tlsListener := tls.NewListener(tcpListener, tlsConfig)

	// Envolver con nuestro wrapper para interceptar errores de handshake
	wrappedListener := &TLSListenerWrapper{Listener: tlsListener}
	defer wrappedListener.Close()

	// Crear servidor
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
		Handler:   http.DefaultServeMux,
	}

	// Servir con listener personalizado
	if err := server.Serve(wrappedListener); err != nil {
		log.Fatalf("Error en servidor: %v", err)
	}
}
