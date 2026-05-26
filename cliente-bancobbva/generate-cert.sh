#!/bin/bash 
set -e
echo "Generando la clave privada para el banco cliente BBVA..."
openssl genpkey -algorithm RSA -out bbva-key.pem
echo "Generando la solicitud de firma (CSR)..."
openssl req -new -key bbva-key.pem -out bbva-req.pem -subj "/C=AR/ST=Buenos Aires/L=CABA/O=BBVA/CN=BancoBBVA"
echo "CSR del banco BBVA creada"

