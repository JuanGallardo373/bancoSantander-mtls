#!/bin/bash 
set -e
echo "Generando la clave privada para el cliente mercadopago..."
openssl genpkey -algorithm RSA -out mpago-key.pem
echo "Generando la solicitud de firma (CSR)..."
openssl req -new -key mpago-key.pem -out mpago-req.pem -subj "/C=AR/ST=Buenos Aires/L=CABA/O=MercadoPago/CN=mercadopago"
echo "CSR de mercadopago creada"
