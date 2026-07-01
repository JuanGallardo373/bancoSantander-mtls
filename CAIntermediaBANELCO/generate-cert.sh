#!/bin/bash
set -e
mkdir -p private newcerts certs
touch index.txt
echo 1000 > serial
echo "Generando clave privada y CSR de BANELCO (CA Intermedia)..."
openssl req -new -nodes -newkey rsa:2048 \
    -keyout banelco-inter.key -out banelco-inter.csr \
    -config ../CABancoCentral/ca.conf \
    -subj "/C=AR/ST=Buenos Aires/L=CABA/O=BANELCO SA/CN=BanelcoCAIntermedia"
echo "Solicitud de firma creada correctamente"