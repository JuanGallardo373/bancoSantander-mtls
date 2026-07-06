#!/bin/bash
# Este script crea el certificado y la llave privada para el server del banco santander
set -e
mkdir -p certs
echo "Generando la clave privada para el servidor del Banco Santader..."
openssl genpkey -algorithm RSA -out certs/santander.key
echo "Generando la solicitud de firma (CSR)..."
openssl req -new -key certs/santander.key -out certs/santander.csr -subj "/C=AR/ST=Buenos Aires/L=CABA/O=Santander/CN=localhost"
echo "Solicitud de firma creada correctamente"
