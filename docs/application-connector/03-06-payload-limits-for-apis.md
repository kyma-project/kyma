---
title: Payload size limits for registering APIs
type: Details
---

The Application Connector allows you to adjust the payload size limit for registering API definitions. You can tune the limit individually for every Application in your Kyma cluster.

The `nginx.ingress.kubernetes.io/proxy-body-size` annotation defines the maximum payload size. By default, every Application you create comes with the payload size limit set to 5 MB. You can adjust it to fit the needs of your implementation.

To change the maximum payload size for an API definition, edit the configuration of the Ingress of the Application for which you want to tune the limit. Run this command to edit the Ingress configuration:

```
kubectl -n kyma-integration edit ingress {APPLICATION_NAME}-application
```
