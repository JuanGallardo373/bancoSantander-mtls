#!/bin/bash
set -e
echo "Firmando certificados para los bancos y billeteras autorizados..."
openssl ca -config ca.conf -extensions server_extensions -out ../servidor-banco/banco-cert.pem -infiles ../servidor-banco/banco-req.pem
openssl ca -config ca.conf -extensions client_extensions -out ../cliente-mercadopago/mpago-cert.pem -infiles ../cliente-mercadopago/mpago-req.pem
openssl ca -config ca.conf -extensions client_extensions -out ../cliente-bancobbva/bbva-cert.pem -infiles ../cliente-bancobbva/bbva-req.pem
echo "Certificados firmados por el Banco Central (CA)"
