# Certificates

## Overview

Certificates Helm chart is used for managing Kyma certificates.

## Details

Depending on the scenario, this Helm chart creates either:
- a plain k8s Secret - `kyma-gateway-certs`
- a Gardener certificate - `kyma-tls-cert`

If the plain k8s Secret is created, the certificate is either provided, or the default one within the Helm values is used. The default certificate is valid for at least 6 months from the installation date.

### Overrides handling

Users can control the behavior of this chart using three overrides: `global.domainName`, `global.tlsCrt` and `global.tlsKey`
These overrides are optional, but if you use TLS overrides remember to define both of them. If only one TLS override is defined and the other is empty, the chart assumes TLS overrides are **not** provided.

The table summarizes what the generated output is, basing on the overrides provided.

| Domain name | TLS overrides | What is generated |
|--|--|--|
| n/a | provided | a Secret with the user-provided values |
| provided | not provided | a Certificate object configured with the given domain name |
| not provided | not provided | a Secret with a default static value defined in this chart |

## Generate a new default certificate
If the provided self-signed default certificate needs to be updated, use `openssl` and create a new one for the `kyma.example.com` domain. Use the following command:
 ```
openssl req -newkey rsa:3072 -keyout kyma.key -x509 -days 365 -out kyma.example.crt
 ```



