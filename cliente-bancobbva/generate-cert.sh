#!/bin/bash 
set -e
mkdir -p certs
echo "Generando la clave privada y CSR para el banco cliente BBVA..."
openssl req -new -nodes -newkey rsa:2048 \
    -keyout certs/bbva.key -out certs/bbva.csr -subj "/C=AR/ST=Buenos Aires/L=CABA/O=BBVA/CN=BancoBBVA"
echo "CSR del banco BBVA creada"