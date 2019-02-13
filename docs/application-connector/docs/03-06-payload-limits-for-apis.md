---
title: Payload size limits for registering APIs
type: Details
---

The Application Connector allows you to adjust the payload size limit for registering API definitions. You can tune the limit individually for every Application in your Kyma cluster.

To change the maximum payload size for an API definition, edit the configuration of the Ingress of the Application for which you want to tune the limit. Run this command to edit the Ingress configuration:

```
kubectl -n kyma-integration edit ingress {APPLICATION_NAME}-application
```

The maximum payload size is defined by the `nginx.ingress.kubernetes.io/proxy-body-size` annotation. By default, every Application you create comes with the payload limit set to 5 MB. You can adjust the size to fit the needs of your implementation.
