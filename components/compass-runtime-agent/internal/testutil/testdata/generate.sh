#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

openssl genrsa -out ${DIR}/ca.key.pem 2048
openssl req -new -sha256 -key ${DIR}/ca.key.pem -out ${DIR}/ca.csr.pem -subj "/CN=localhost"
openssl req -x509 -sha256 -days 365000 -key ${DIR}/ca.key.pem -in ${DIR}/ca.csr.pem -out ${DIR}/ca.crt.pem

openssl genrsa -out ${DIR}/client.key.pem 2048
openssl req -new -key ${DIR}/client.key.pem -nodes -days 365000 -out ${DIR}/client.csr.pem  -subj "/CN=localhost"
openssl x509 -req -days 36500 -in ${DIR}/client.csr.pem -CA ${DIR}/ca.crt.pem -CAkey ${DIR}/ca.key.pem -out ${DIR}/client.crt.pem -CAcreateserial

cat ${DIR}/client.crt.pem ${DIR}/ca.crt.pem > cert.chain.pem