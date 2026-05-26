package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	Timestamp     time.Time `json:"timestamp"`
	EventType     string    `json:"event_type"`
	ClientIP      string    `json:"client_ip"`
	ClientName    string    `json:"client_name"`
	ErrorMessage  string    `json:"error_message"`
	CertChain     []string  `json:"cert_chain,omitempty"`
	IsExpired     bool      `json:"is_expired,omitempty"`
	IsSelfSigned  bool      `json:"is_self_signed,omitempty"`
}

var (
	logFile      *os.File
	jsonEncoder  *json.Encoder
	transferIDCounter int
)

func init() {
	var err error
	// Crear o abrir el archivo de logs JSON
	logFile, err = os.OpenFile("../logs/anomalies.jsonl", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error al abrir archivo de logs: %v", err)
	}

	// Crear directorio de logs si no existe
	os.MkdirAll("../logs", 0755)
	logFile, err = os.OpenFile("../logs/anomalies.jsonl", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error al abrir archivo de logs: %v", err)
	}

	jsonEncoder = json.NewEncoder(logFile)
}

func logAnomaly(anomaly AnomalyLog) {
	anomaly.Timestamp = time.Now()
	if err := jsonEncoder.Encode(anomaly); err != nil {
		log.Printf("Error escribiendo log de anomalía: %v", err)
	}
	logFile.Sync()
}

func getClientCertInfo(r *http.Request) (string, bool, bool) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return "unknown", false, false
	}

	cert := r.TLS.PeerCertificates[0]
	clientName := cert.Subject.CommonName
	
	// Verificar si el certificado está expirado
	isExpired := time.Now().After(cert.NotAfter)
	
	// Verificar si es autofirmado (Issuer == Subject)
	isSelfSigned := cert.Issuer.String() == cert.Subject.String()

	return clientName, isExpired, isSelfSigned
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtener información del certificado del cliente
	clientName, isExpired, isSelfSigned := getClientCertInfo(r)

	// Loguear anomalías si existen
	if isExpired || isSelfSigned {
		anomaly := AnomalyLog{
			EventType:    "SUSPICIOUS_CERTIFICATE",
			ClientIP:     r.RemoteAddr,
			ClientName:   clientName,
			IsExpired:    isExpired,
			IsSelfSigned: isSelfSigned,
		}

		if isExpired {
			anomaly.ErrorMessage = "Certificado cliente expirado detectado"
		}
		if isSelfSigned {
			anomaly.ErrorMessage = anomaly.ErrorMessage + "; Certificado autofirmado detectado"
		}

		logAnomaly(anomaly)
		log.Printf("⚠️  ANOMALÍA: %s | Cliente: %s | IP: %s", anomaly.ErrorMessage, clientName, r.RemoteAddr)
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

func main() {
	defer logFile.Close()

	// Cargar certificados de cliente (CA)
	caCert, err := os.ReadFile("./certs/ca-cert.pem")
	if err != nil {
		log.Fatalf("Error cargando CA certificate: %v", err)
	}

	caCertPool := tls.NewCertPool()
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

	// Configurar TLS con mTLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
	}

	// Configurar servidor HTTPS
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
	}

	// Registrar handlers
	http.HandleFunc("/transfer", transferHandler)
	http.HandleFunc("/health", healthHandler)

	log.Println("🚀 Servidor Banco Santander iniciado en https://localhost:8443")
	log.Println("📋 Esperando solicitudes con mTLS...")
	log.Println("📊 Logs de anomalías guardados en: ../logs/anomalies.jsonl")

	if err := server.ListenAndServeTLS("./certs/servidor-cert.pem", "./certs/servidor-key.pem"); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
