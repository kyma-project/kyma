---
title: Ingress and Egress traffic
---

## Ingress

Kyma uses the [Istio Ingress Gateway](https://istio.io/latest/docs/reference/config/networking/gateway/) to handle all incoming traffic, manage TLS termination, and handle mTLS communication between the cluster and external services. By default, the [`kyma-gateway`](https://github.com/kyma-project/kyma/blob/main/resources/certificates/templates/gateway.yaml) configuration defines the points of entry to expose all applications using the supplied domain and certificates.
Applications are exposed using the [API Gateway](../../01-overview/main-areas/api-exposure/apix-01-api-gateway.md) controller.

The configuration specifies the following parameters and their values:

| Parameter | Description | Value|
|-----| ---| -----|
| **spec.servers.port** | The ports gateway listens on.  Port `80` is automatically redirected to `443`.| `443`, `80`.|
| **spec.servers.tls.minProtocolVersion** | The minimum protocol version required by the TLS connection. | `TLSV1_2` protocol version. `TLSV1_0` and `TLSV1_1` are rejected. |
| **spec.servers.tls.cipherSuites** | Accepted cypher suites. | `ECDHE-RSA-CHACHA20-POLY1305`, `ECDHE-RSA-AES256-GCM-SHA384`, `ECDHE-RSA-AES256-SHA`, `ECDHE-RSA-AES128-GCM-SHA256`, `ECDHE-RSA-AES128-SHA`|

## TLS management

Kyma employs the Bring Your Own Domain/Certificates model that requires you to supply the certificate and key during installation. You can do it using a [values.yaml](../operations/03-change-kyma-config-values.md) file. See a sample YAML file specifying the values to override:

```yaml
global:
  tlsCrt: "CERT"
  tlsKey: "CERT_KEY"
```

During installation, the values are propagated in the cluster to all components that require them.

### Gardener

You can install Kyma on top of [Gardener](https://gardener.cloud/) managed instances that use the already present certificate management service. Because Gardener creates the certificates or domain, you don't have to provide them. The certificate is a Custom Resource managed by Gardener, and is a wildcard certificate for the whole domain.

## Egress

Currently no Egress limitations are implemented, meaning that all applications deployed in the Kyma cluster can access outside resources without limitations.

>**NOTE:** In the case of connection problems with external services, it may be required to create an [Service Entry](https://istio.io/latest/docs/reference/config/networking/service-entry/) object to register the service.
