# Certificates

## Overview

Certificates Helm chart is used for managing Kyma certificates. 

## Details

Depending on the scenario, this Helm chart creates the following secrets:

- `kyma-gateway-certs`
- `apiserver-proxy-tls-cert`
- optional `ingress-tls-cert`, which contains a copy of the TLS certificate from the `kyma-gateway-certs` Secret. It is only created if the certificate is self-signed.

Each job in this chart handles a different scenario:

- job-legacy - Support for the old way of managing certificates. User provides TLS Key and Certificate by overrides `global.tlsKey` and `global.tlsCrt`. Job puts them into secrets.
- job-gardener - This scenario is used when working on Gardener. Certificate CRs are created, which generate Secrets with TLS Certificate and private key.

## How to generate new default certificate
If no certificate is provided during installation of Kyma, there is a default certificate provided within the
chart values of the certificate helm chart.
The certificate is installed as a fallback for created for the sample domain kyma.example.com
It was created as a self signed certificate using openssl.
If the certificate is outdated, a new one can be generated via doing following steps:

 ```
openssl req -newkey rsa:3072 -keyout kyma.key -x509 -days 365 -out kyma.example.crt
 ```