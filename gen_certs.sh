#!/usr/bin/env zsh

openssl genrsa -out rootCA.key 4096
openssl req -new -x509 -days 365 -key rootCA.key -out rootCA.crt -subj "/C=AU/ST=Some-State/O=test/CN=localhost"
openssl genrsa -out localhost.key 2048
openssl req -new -key localhost.key -out localhost.csr -subj "/O=test/CN=localhost"
openssl x509 -req -in localhost.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out localhost.crt -days 365 -extensions v3_req -extfile <(echo "[v3_req]"; echo "subjectAltName=DNS:localhost")
rm localhost.csr
