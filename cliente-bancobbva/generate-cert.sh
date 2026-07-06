#!/bin/bash 
set -e
mkdir -p certs
echo "Generando la clave privada para el banco cliente BBVA..."
openssl genpkey -algorithm RSA -out certs/bbva.key
echo "Generando la solicitud de firma (CSR)..."
openssl req -new -key certs/bbva.key -out certs/bbva.csr -subj "/C=AR/ST=Buenos Aires/L=CABA/O=BBVA/CN=BancoBBVA"
echo "CSR del banco BBVA creada"