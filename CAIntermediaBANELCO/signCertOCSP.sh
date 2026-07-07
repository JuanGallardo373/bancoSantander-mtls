#!/bin/bash
echo "Firmando el certificado para el OCSP Responder de BANELCO..."
openssl ca -config ../CABancoCentral/ca.conf -extensions ocsp_extensions \
    -cert banelco-inter.crt -keyfile banelco-inter.key \
    -out ocsp.crt \
    -infiles ocsp.csr
echo "Certificado firmado correctamente"
