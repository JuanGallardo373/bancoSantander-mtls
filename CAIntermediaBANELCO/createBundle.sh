cat banelco-inter.crt ../CABancoCentral/bcra-raiz.crt > bundle.crt
echo "Bundle de certificados creado correctamente: bundle.crt"
echo "Contenido del bundle:"
cat bundle.crt