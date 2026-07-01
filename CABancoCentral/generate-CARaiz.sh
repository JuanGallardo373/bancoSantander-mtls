#!/bin/bash
set -e
mkdir -p private newcerts certs
touch index.txt
echo 1000 > serial
echo "Generando clave privada y certificado del Banco Central (CA)..."
openssl req -new -x509 -days 7300 -nodes -newkey rsa:4096 -extensions v3_ca -keyout private/bcra-raiz.key \
	-out bcra-raiz.crt -config ca.conf -subj "/C=AR/ST=Buenos Aires/L=CABA/O=BCRA/CN=BancoCentralCA"
echo ""
echo "Información del certificado:"
openssl x509 -in bcra-raiz.crt -text -noout | grep -E "Subject:|Issuer:|Not Before|Not After|Public-Key"
