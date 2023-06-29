---
title: Ingress and Egress traffic
---

## Ingress

Kyma uses the [Istio Ingress Gateway](https://istio.io/latest/docs/reference/config/networking/gateway/) to handle all incoming traffic, manage TLS termination, and handle mTLS communication between the cluster and external services. By default, the [`kyma-gateway`](https://github.com/kyma-project/kyma/blob/main/resources/certificates/templates/gateway.yaml) configuration defines the points of entry to expose all applications using the supplied domain and certificates.
Applications are exposed using the [API Gateway](../../01-overview/api-exposure/apix-01-api-gateway.md) controller.

The configuration specifies the following parameters and their values:

| Parameter | Description | Value|
|-----| ---| -----|
| **spec.servers.port** | The ports gateway listens on.  Port `80` is automatically redirected to `443`.| `443`, `80`.|
| **spec.servers.tls.minProtocolVersion** | The minimum protocol version required by the TLS connection. | `TLSV1_2` protocol version. `TLSV1_0` and `TLSV1_1` are rejected. |
| **spec.servers.tls.cipherSuites** | Accepted cypher suites. | `ECDHE-RSA-CHACHA20-POLY1305`, `ECDHE-RSA-AES256-GCM-SHA384`, `ECDHE-RSA-AES256-SHA`, `ECDHE-RSA-AES128-GCM-SHA256`, `ECDHE-RSA-AES128-SHA`|

## TLS management

Kyma employs the Bring Your Own Domain/Certificates model that requires you to supply the domain, certificate, and key during installation. Read the tutorial to learn how to [set up or update your custom domain TLS certificate in Kyma](../../03-tutorials/00-security/sec-01-tls-certificates-security.md).

If you don't want to use your custom certificate, you can choose between a self-signed certificate or one managed by the Gardener [Certificate Management](https://github.com/gardener/cert-management) component.

### Self-signed certificate

If you deploy Kyma on a local k3d cluster, you are issued with a certificate for the `local.kyma.dev` cluster domain. The certificate is valid for about 10 years.

If you deploy Kyma on a remote non-Gardener cluster, you are issued with a certificate for the `kyma.example.com` cluster domain. The certificate is valid for at least 6 months.

The self-signed certificates are neither managed nor rotated.

### Gardener Certificate Management

If you use a Gardener cluster or have the Certificate Management component installed on any environment, you are issued with a certificate that is rotated and managed by the Certificate Management component.

The prerequisite is having the [Certificate Management](https://github.com/gardener/cert-management) and [External DNS Management](https://github.com/gardener/external-dns-management) components installed and properly configured. The Certificate Management uses the default issuer for generating certificates.

Once you deploy Kyma on a Gardener cluster, Kyma CLI reads the pre-configured domain and uses it for issuing the certificate. If you use a non-Gardener cluster, you have to provide the domain as an override during the deployment process. The Certificate is a custom resource managed by Gardener, and is a wildcard certificate for the whole domain. The certificate's validity depends on the configuration details. For more information, read the official Gardener documentation.

## Egress

Currently no Egress limitations are implemented, meaning that all applications deployed in the Kyma cluster can access outside resources without limitations.

>**NOTE:** In the case of connection problems with external services, it may be required to create an [Service Entry](https://istio.io/latest/docs/reference/config/networking/service-entry/) object to register the service.
