#!/bin/bash
# Este script crea el certificado y la llave privada para el server del banco santander
set -e
mkdir -p certs
echo "Generando la clave privada y CSR para el servidor del Banco Santander..."
openssl req -new -nodes -newkey rsa:2048 \
    -keyout certs/santander.key -out certs/santander.csr -subj "/C=AR/ST=Buenos Aires/L=CABA/O=Santander/CN=localhost"
echo "Solicitud de firma creada correctamente"
