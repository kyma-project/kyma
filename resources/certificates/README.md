# Certificates

## Overview

Certificates Helm chart is used for managing Kyma certificates. 

## Details

Depending on the scenario, this Helm chart creates secrets:

- `kyma-gateway-certs`
- `apiserver-proxy-tls-cert`
- optional `ingress-tls-cert`, which contains a copy of the TLS certificate from the `kyma-gateway-certs` Secret. It is only created if the certificate is self-signed.

Each job in this chart handles a different scenario:

- job-user-provided - Requires cert-manager installed on the cluster. User provides a ClusterIssuer CR. Then, Certificate CRs are created, which generate a Secret with TLS Certificate and private key based on the provided ClusterIssuer.