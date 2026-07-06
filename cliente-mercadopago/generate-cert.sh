#!/bin/bash 
set -e
mkdir -p certs
echo "Generando la clave privada para el cliente mercadopago..."
openssl genpkey -algorithm RSA -out certs/mpago.key
echo "Generando la solicitud de firma (CSR)..."
openssl req -new -key certs/mpago.key -out certs/mpago.csr -subj "/C=AR/ST=Buenos Aires/L=CABA/O=MercadoPago/CN=mercadopago"
echo "CSR de mercadopago creada"
