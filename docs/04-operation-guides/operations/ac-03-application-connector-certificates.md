---
title: Provide a custom Application Connector certificate and key
---

Application Connector is secured with a client certificate verified by the Istio Ingress Gateway.

The root CA certificates are generated and stored as Kubernetes Secrets by Application Connector Certs Setup job.

By default, the server key and certificate are automatically generated.
You can provide a custom server certificate and key during the installation by overriding these default values:

>**NOTE:** To use a custom certificate and key, you must override both values. If either the certificate or key is incorrect or isn't provided, a new certificate and key pair is generated.

```yaml
global.applicationConnectorCaKey: "{BASE64_ENCODED_PRIVATE_KEY}"
global.applicationConnectorCa: "{BASE64_ENCODED_CERTIFICATE}"
```

>**TIP:** Read more about how to [change Kyma settings](03-change-kyma-config-values.md).
