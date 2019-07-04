---
title: Application Connector Certificates  
type: Details
---

Application Connector is secured with the client certificate which is being verified by the Istio Ingress Gateway.

By default, the server key and certificate are automatically generated. 
The user can provide his own server certificate and key during the installation by setting them as the following overrides:
```
global.applicationConnectorCaKey: {BASE64_ENCODED_PRIVATE_KEY}
global.applicationConnectorCa: {BASE64_ENCODED_CERTIFICATE}
```

The example Config Map containing the overrides:
```
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

>**NOTE:** For the Application Connector to use provided key and certificate, both values need to be specified. In case any of the values are missing or is incorrect, the new certificate and key pair will be generated.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents: 
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)

Certificates are generated and put to Kubernetes Secrets by the Application Connector Certs Setup Job.

