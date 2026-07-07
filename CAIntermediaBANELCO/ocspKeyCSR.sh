#!/bin/bash
echo "Generando la clave privada y CSR para el OCSP Responder de BANELCO..."
openssl req -new -nodes -newkey rsa:2048 \
    -keyout ocsp.key -out ocsp.csr \
    -config ../CABancoCentral/ca.conf \
    -subj "/C=AR/ST=Buenos Aires/L=CABA/O=BANELCO SA/CN=OCSP Responder BANELCO"
echo "Solicitud de firma creada correctamente"
