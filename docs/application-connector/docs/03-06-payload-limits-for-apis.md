---
title: Payload limits for registering APIs
type: Details
---

The Application Connector allows you to register API definitions that are up to 5 mb in size. You can tune the payload limit individually for every Application in your Kyma cluster.

To change the maximum payload size for an API definition, edit the configuration of the Ingress of the Application for which you want to tune the limit. Run this command to edit the Ingress configuration:

```
kubectl -n kyma-integration edit ingress {APPLICATION_NAME}-application
```

The maximum payload size is defined by the `nginx.ingress.kubernetes.io/proxy-body-size` annotation. By default, every Application you create comes with the payload limit set to 5mb. You can set the size to fit the needs of your implementation.

>**NOTE:** You can set the payload size limit to any value between 0,1 and 5 mb. 
