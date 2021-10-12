# Certificates

## Overview

Certificates Helm chart is used for managing Kyma certificates. 

## Details

Depending on the scenario, this Helm chart creates either:
- a plain k8s Secret - `kyma-gateway-certs`
- a Gardener certificate - `kyma-tls-cert`

If the plain k8s Secret is created, the certificate is either provided, or the default one within the Helm values is used. The default certificate is valid for at least 6 months from the installation date.


## Generate a new default certificate
If the provided self-signed default certificate needs to be updated, use `openssl` and create a new one for the `kyma.example.com` domain.
Use following command:
 ```
openssl req -newkey rsa:3072 -keyout kyma.key -x509 -days 365 -out kyma.example.crt
 ```