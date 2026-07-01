openssl ca -config ../CABancoCentral/ca.conf -database index.txt \
    -cert banelco-inter.crt -keyfile banelco-inter.key \
    -revoke ../cliente-mercadopago/certs/mpago.crt