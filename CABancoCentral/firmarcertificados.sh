#!/bin/bash
set -e
echo "Firmando certificados para el servidor del Banco Santander y los bancos y Fintech clientes..."
openssl ca -config ca.conf -extensions server_extensions \
    -out ../servidor-banco/certs/santander.crt -infiles ../servidor-banco/certs/santander.csr
openssl ca -config ca.conf -extensions client_extensions \
    -out ../cliente-mercadopago/certs/mpago.crt -infiles ../cliente-mercadopago/certs/mpago.csr
openssl ca -config ca.conf -extensions client_extensions \
    -out ../cliente-bancobbva/certs/bbva.crt -infiles ../cliente-bancobbva/certs/bbva.csr
echo "Certificados firmados por el Banco Central (CA)"
