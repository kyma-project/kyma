# Istio Ingress Gateway

The Application Connector module relies on Istio Ingress Gateway as an endpoint for incoming requests from external systems.

The Kyma Istio module is pre-installed by default on a managed Kyma runtime. Alternatively, a [custom Istio installation](https://kyma-project.io/#/02-get-started/01-quick-install) is also possible.

Istio Gateway is created during the Application Connector module installation.

## DNS Name

The DNS name of Ingress Gateway is cluster-dependent. For the managed Kyma clusters, it follows the `gateway.{cluster-dns}` format.

## Security

### Client Certificates

For external systems automatically integrated by Unified Customer Landscape (UCL), the Application Connector module uses the mutual TLS protocol (mTLS) with client authentication enabled. As a result, anyone attempting to connect to the Application Connector module must present a valid client certificate dedicated to a specific Application. This way, the traffic is fully encrypted, and the client has a valid identity.

### TLS Certificate Verification

By default, the TLS certificate verification is enabled when you send data and requests to any application.
You can [disable the TLS certificate verification](../tutorials/01-50-disable-tls-certificate-verification.md) in the communication between Kyma and an application to allow Kyma to send requests and data to an unsecured application. Disabling the certificate verification can be useful in certain testing scenarios.
