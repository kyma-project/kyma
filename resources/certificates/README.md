# Certificates

## Overview

Certificates Helm chart is used for managing Kyma certificates. 

## Details

Depending on the scenario, this Helm chart creates either:
- plain k8s secret   `kyma-gateway-certs`
- gardener certificate `kyma-tls-cert`

For the plain k8s secret case the certificate is either provided or the default one within the Helm values
will be used, which will be valid for at least 6 months after installation.


## How to generate new default certificate
If the provided self signed default certificate needs to be updated, please use `openssl` and create for 
domain `kyma.example.com`
Use following command:
 ```
openssl req -newkey rsa:3072 -keyout kyma.key -x509 -days 365 -out kyma.example.crt
 ```