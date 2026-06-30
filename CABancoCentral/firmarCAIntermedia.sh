#!/bin/bash
set -e
echo "Banco Central (CA Raiz) firmando certificado de BANELCO (CA Intermedia)..."
openssl ca -config ca.conf -extensions v3_intermedia \
    -days 3650 \
    -cert bcra-raiz.crt -keyfile private/bcra-raiz.key \
    -out ../CAIntermediaBANELCO/banelco-inter.crt -infiles ../CAIntermediaBANELCO/banelco-inter.csr
echo "Certificado firmado por el Banco Central (CA)"