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
	Timestamp       time.Time `json:"timestamp"`
	EventType       string    `json:"event_type"`
	ClientIP        string    `json:"client_ip"`
	ClientName      string    `json:"client_name"`
	ErrorMessage    string    `json:"error_message"`
	CertChain       []string  `json:"cert_chain,omitempty"`
	IsExpired       bool      `json:"is_expired,omitempty"`
	IsSelfSigned    bool      `json:"is_self_signed,omitempty"`
	HandshakeError  bool      `json:"handshake_error,omitempty"`
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
	logFile, err = os.OpenFile("../logs/access_anomalies.jsonl", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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

// getClientCertInfo extrae información del certificado del cliente
func getClientCertInfo(r *http.Request) (string, bool, bool, []string) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return "unknown", false, false, nil
	}

	cert := r.TLS.PeerCertificates[0]
	clientName := cert.Subject.CommonName

	// Verificar si el certificado está expirado
	isExpired := time.Now().After(cert.NotAfter)

	// Verificar si es autofirmado (Issuer == Subject)
	isSelfSigned := cert.Issuer.String() == cert.Subject.String()

	// Construir cadena de certificados
	certChain := []string{
		fmt.Sprintf("Subject: %s", cert.Subject.String()),
		fmt.Sprintf("Issuer: %s", cert.Issuer.String()),
		fmt.Sprintf("NotBefore: %s", cert.NotBefore.String()),
		fmt.Sprintf("NotAfter: %s", cert.NotAfter.String()),
		fmt.Sprintf("SerialNumber: %s", cert.SerialNumber.String()),
	}

	return clientName, isExpired, isSelfSigned, certChain
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtener información del certificado del cliente
	clientName, isExpired, isSelfSigned, certChain := getClientCertInfo(r)

	// Loguear anomalías si existen
	if isExpired || isSelfSigned {
		anomaly := AnomalyLog{
			EventType:    "SUSPICIOUS_CERTIFICATE",
			ClientIP:     extractClientIP(r.RemoteAddr),
			ClientName:   clientName,
			IsExpired:    isExpired,
			IsSelfSigned: isSelfSigned,
			CertChain:    certChain,
		}

		if isExpired {
			anomaly.ErrorMessage = "Certificado cliente expirado detectado"
		}
		if isSelfSigned {
			if anomaly.ErrorMessage != "" {
				anomaly.ErrorMessage = anomaly.ErrorMessage + "; "
			}
			anomaly.ErrorMessage = anomaly.ErrorMessage + "Certificado autofirmado detectado"
		}

		logAnomaly(anomaly)
		log.Printf("⚠️  ANOMALÍA: %s | Cliente: %s | IP: %s",
			anomaly.ErrorMessage, clientName, anomaly.ClientIP)
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

// TLSConnWrapper intercepta el handshake TLS
type TLSConnWrapper struct {
	net.Conn
	wrapped bool
}

func (c *TLSConnWrapper) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)

	// Si hay error en el primer Read (handshake), registrarlo
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
	caCert, err := os.ReadFile("../CABancoCentral/cacert.pem")
	if err != nil {
		log.Fatalf("Error cargando CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatal("Error agregando CA certificate al pool")
	}

	// Cargar certificado y clave del servidor
	serverCert, err := tls.LoadX509KeyPair(
		"./banco-cert.pem",
		"./banco-key.pem",
	)
	if err != nil {
		log.Fatalf("Error cargando certificado del servidor: %v", err)
	}

	// Configurar TLS con mTLS
	// ClientAuth: RequireAndVerifyClientCert requiere que el cliente presente un certificado válido
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	// Registrar handlers
	http.HandleFunc("/transfer", transferHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("🚀 Servidor Banco Santander iniciado en https://localhost:8443")
	log.Println("📋 Esperando solicitudes con mTLS obligatorio...")
	log.Println("📊 Logs de anomalías guardados en: ../logs/access_anomalies.jsonl")
	log.Println()
	log.Println("🔐 Configuración mTLS:")
	log.Println("   ✓ Requiere certificado cliente")
	log.Println("   ✓ Verifica firma de CA")
	log.Println("   ✓ Rechaza certificados autofirmados")
	log.Println("   ✓ Rechaza certificados expirados")
	log.Println("   ✓ Registra todos los errores en anomalies.jsonl")
	log.Println()

	// Crear listener TCP
	tcpListener, err := net.Listen("tcp", ":8443")
	if err != nil {
		log.Fatalf("Error creando listener TCP: %v", err)
	}
	defer tcpListener.Close()

	// Envolver con TLS
	tlsListener := tls.NewListener(tcpListener, tlsConfig)
	
	// Envolver con nuestro wrapper para interceptar errores
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


