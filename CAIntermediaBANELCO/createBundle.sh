cat banelco-inter.crt ../CABancoCentral/bcra-raiz.pem > bundle.crt
echo "Bundle de certificados creado correctamente: bundle.crt"
echo "Contenido del bundle:"
cat bundle.crt