---
title: Application Connector Certificates  
type: Details
---

Application Connector is secured with the client certificate which is being verified by the Istio Ingress Gateway.

By default the server key and certificate are automatically generated. 
The user can provide his own server certificate and key during the installation by setting them as the following overrides:
```
global.applicationConnectorCaKey: {PRIVATE_KEY}
global.applicationConnectorCa: {CERTIFICATE}
```

>**NOTE:** For the Application Connector to use provided key and certificate, both values need to be specified.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents: 
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)

Certificates are generated and put to Kubernetes Secrets by the Application Connector Certs Setup Job.

