# Certificates

## Overview

Certificates Helm chart is used for managing Kyma certificates.

## Details

Depending on the scenario, this Helm chart creates either:
- a plain k8s Secret - `kyma-gateway-certs`
- a Gardener certificate - `kyma-tls-cert`

If the plain k8s Secret is created, the certificate is either provided, or the default one within the Helm values is used. The default certificate is valid for at least 6 months from the installation date.

### Overrides handling

Users can control the behavior of this chart via three overrides: `global.domainName`, `global.tlsCrt` and `global.tlsKey`
These overrides are optional, but the TLS overrides must be both defined, if used. If only one TLS override is defined and the other is empty, the chart assumes TLS overrides are **not** provided.

The table below summarizes the logic that decides if a plain k8s Secret or a Certificate object is generated, based on provided overrides.

| Domain name | TLS overrides | What is generated |
|--|--|--|
| n/a | provided | Secret with user-provided values |
| provided | not provided | Certificate object configured with given domain name |
| not provided | not provided | Secret with a default static value defined in this chart|

## Generate a new default certificate
If the provided self-signed default certificate needs to be updated, use `openssl` and create a new one for the `kyma.example.com` domain.
Use following command:
 ```
openssl req -newkey rsa:3072 -keyout kyma.key -x509 -days 365 -out kyma.example.crt
 ```



