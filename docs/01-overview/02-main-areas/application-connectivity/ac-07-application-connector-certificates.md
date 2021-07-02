---
title: Application Connector Certificates
---

Application Connector is secured with a client certificate verified by the Istio Ingress Gateway.

The Certificates are generated and stored as Kubernetes Secrets by Application Connector Certs Setup job.

By default, the server key and certificate are automatically generated.
You can provide a custom server certificate and key during the installation by overriding these default values:
```yaml
global.applicationConnectorCaKey: "{BASE64_ENCODED_PRIVATE_KEY}"
global.applicationConnectorCa: "{BASE64_ENCODED_CERTIFICATE}"
```

>**NOTE:** To use a custom certificate and key, you must override both the values. If either the certificate or key is incorrect or isn't provided, a new certificate and key pair is generated.

This is a sample ConfigMap that overrides the default values with a custom certificate and key:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: application-connector-certificate-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.applicationConnectorCa: "{BASE64_ENCODED_CERTIFICATE}"
  global.applicationConnectorCaKey: "{BASE64_ENCODED_PRIVATE_KEY}"
```

>**TIP:** Read more about how to [change Kyma settings](../../../04-operation-guides/operations/03-change-kyma-config-values.md).