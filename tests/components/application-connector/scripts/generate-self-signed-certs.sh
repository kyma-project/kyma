#!/usr/bin/env bash

if [ $# -ne 2 ]; then
  echo "Usage: generate-self-signed-certs.sh <domain> <DIR>"
  exit 1
fi

export APP_URL=$1
export GATEWAY_TEST_CERTS_DIR=$2
export SUBJECT="/C=PL/ST=A/O=SAP/CN=$APP_URL"

echo "Generating certificate for domain: $APP_URL"
openssl version
openssl req -newkey rsa:2048 -nodes -x509 -days 365 -out $GATEWAY_TEST_CERTS_DIR/ca.crt -keyout $GATEWAY_TEST_CERTS_DIR/ca.key -subj $SUBJECT

openssl genrsa -out $GATEWAY_TEST_CERTS_DIR/server.key 2048
openssl genrsa -out $GATEWAY_TEST_CERTS_DIR/client.key 2048

openssl req -new \
  -key $GATEWAY_TEST_CERTS_DIR/server.key \
  -subj $SUBJECT \
  -reqexts SAN \
  -config <(cat /etc/ssl/openssl.cnf \
      <(printf "\n[SAN]\nsubjectAltName=DNS:$APP_URL")) \
  -out $GATEWAY_TEST_CERTS_DIR/server.csr

	openssl x509 -req -days 365 -CA $GATEWAY_TEST_CERTS_DIR/ca.crt -CAkey $GATEWAY_TEST_CERTS_DIR/ca.key -CAcreateserial \
  	-extensions SAN \
  	-extfile <(cat /etc/ssl/openssl.cnf \
    <(printf "\n[SAN]\nsubjectAltName=DNS:$APP_URL")) \
  	-in $GATEWAY_TEST_CERTS_DIR/server.csr -out $GATEWAY_TEST_CERTS_DIR/server.crt

openssl req -new \
  -key $GATEWAY_TEST_CERTS_DIR/client.key \
  -subj $SUBJECT \
  -reqexts SAN \
  -config <(cat /etc/ssl/openssl.cnf \
      <(printf "\n[SAN]\nsubjectAltName=DNS:$APP_URL")) \
  -out $GATEWAY_TEST_CERTS_DIR/client.csr

	openssl x509 -req -days 365 -CA $GATEWAY_TEST_CERTS_DIR/ca.crt -CAkey $GATEWAY_TEST_CERTS_DIR/ca.key -CAcreateserial \
  	-extensions SAN \
  	-extfile <(cat /etc/ssl/openssl.cnf \
    <(printf "\n[SAN]\nsubjectAltName=DNS:$APP_URL")) \
  	-in $GATEWAY_TEST_CERTS_DIR/client.csr -out $GATEWAY_TEST_CERTS_DIR/client.crt