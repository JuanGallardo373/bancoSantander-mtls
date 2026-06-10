#!/bin/bash
set -e
echo "Firmando certificados para los bancos y billeteras autorizados..."
openssl ca -config ca.conf -extensions server_extensions \ 
    -out ../servidor-banco/certs/banco-cert.pem -infiles ../servidor-banco/certs/banco-req.pem
openssl ca -config ca.conf -extensions client_extensions \ 
    -out ../cliente-mercadopago/certs/mpago-cert.pem -infiles ../cliente-mercadopago/certs/mpago-req.pem
openssl ca -config ca.conf -extensions client_extensions \ 
    -out ../cliente-bancobbva/certs/bbva-cert.pem -infiles ../cliente-bancobbva/certs/bbva-req.pem
echo "Certificados firmados por el Banco Central (CA)"
