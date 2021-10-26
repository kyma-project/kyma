---
title: Certificates in Kyma
---

In Kyma, there are 4 types of TLS certificates. The certificates differ, depending on how you install Kyma.

All the certificates store a plain k8s Secret, are created in the `istio-system` Namespace, and are used with the `kyma-gateway-certs` Kyma Gateway.

| Installation type | Certificate type | Validity | Description |
|-------------------|------------------|----------|-------------|
| local k3d | name - we can find it | 10 years but ask Johannes | For local installation, the cluster domain is `local.kyma.dev` and there is no certificate management nor rotation. |
| remote non-Gardener | self-signed certificate | at least 6 months | For remote non-Gardener installation, the cluster domain is `kyma.example.com`, and there is no certificate management nor rotation. |
| remote Gardener | `kyma-tls-cert` | to be checked | For remote Gardener installation, the cluster domain is...  and the certificate is managed and rotated. 
| all | custom domain certificate | custom | For any of the installation types, the user can provide their custom certificate. There is no certificate management nor rotation. See the document to learn how to [set up your custom TLS certificate](../../03-tutorials/sec-01-tls-certificates-security.md).
