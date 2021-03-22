---
title: Application Connector Certificates  
type: Details
---

The Application Connector is secured with a client certificate verified by the Istio Ingress Gateway.

The Certificates are generated and stored as Kubernetes Secrets by the Application Connector Certs Setup job.

By default, the server key and certificate are automatically generated. 
You can provide a custom server certificate and key during the installation by setting them as the following overrides:
```yaml
global.applicationConnectorCaKey: "{BASE64_ENCODED_PRIVATE_KEY}"
global.applicationConnectorCa: "{BASE64_ENCODED_CERTIFICATE}"
```

>**NOTE:** To use a custom certificate and key, you must provide both values through overrides. If either the certificate or key is incorrect or isn't provided, a new certificate and key pair is generated.

This is a sample ConfigMap that contains overrides with a custom certificate and key:
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

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents: 
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)

