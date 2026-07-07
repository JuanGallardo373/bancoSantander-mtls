#!/bin/bash
echo "Iniciando el OCSP Responder en el puerto 2560..."
openssl ocsp -index index.txt -port 2560 \
    -rsigner ocsp.crt -rkey ocsp.key \
    -CA banelco-inter.crt -text
echo "OCSP Responder iniciado en el puerto 2560"
