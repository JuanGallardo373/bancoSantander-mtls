#!/bin/bash 
set -e
mkdir -p certs
echo "Generando la clave privada y CSR para el cliente mercadopago..."
openssl req -new -nodes -newkey rsa:2048 \
    -keyout certs/mpago.key -out certs/mpago.csr -subj "/C=AR/ST=Buenos Aires/L=CABA/O=Mercado Pago/CN=mercadopago"
echo "CSR de mercadopago creada"
