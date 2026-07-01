echo "Firmando el certificado para el OCSP Responder de BANELCO..."
openssl ca -config ../CABancoCentral/ca.conf -extensions ocsp_extensions \
    -cert banelco-inter.crt -keyfile banelco-inter.key \
    -out ocsp.crt \
    -infiles ocsp.csr
echo "Certificado firmado correctamente"
echo "Iniciando el OCSP Responder en el puerto 2560..."
openssl ocsp -index index.txt -port 2560 \
    -rsigner ocsp.crt -rkey ocsp.key \
    -CA banelco-inter.crt -text
echo "OCSP Responder iniciado en el puerto 2560"