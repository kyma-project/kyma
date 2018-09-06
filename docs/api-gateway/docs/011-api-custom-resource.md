---
title: Api
type: Custom Resource
---

The `api.gateway.kyma.cx` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format the API Controller listens for. To get the up-to-date CRD and show
the output in the `yaml` format, run this command:
```
kubectl get crd apis.gateway.kyma.cx -o yaml
```

## Sample Custom Resource

This is a sample CR that the API-Controller listens for to expose a service. This example has the **authentication** section specified which makes the API Controller create an Istio Authentication Policy for this service.

```
apiVersion: gateway.kyma.cx/v1alpha2
kind: api
metadata:
    name: sample-api
spec:
    service:
      name: kubernetes
      port: 443
    hostname: kubernetes.kyma.local
    authentication:
    - type: JWT
      jwt:
        issuer: https://accounts.google.com
        jwksUri: https://www.googleapis.com/oauth2/v3/certs
```

This table lists all the possible parameters of a given resource together with their descriptions:


| Field   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the exposed API |
| **service.name**, **service.port** | **YES** | Specifies the name and the communication port of the exposed service. |
| **spec.hostname** | **YES** | Specifies the service's external inbound communication address. |
| **spec.authentication** | **NO** | Allows to specify an array of authentication policies that secure the service. |
| **authentication.type** | **YES** | Specifies the desired authentication method that secures the exposed service. |
| **authentication.jwt.issuer**, **authentication.jwt.jwksUri** | **YES** | Specifies the issuer of the tokens used to access the services, as well as the JWKS endpoint URI. |
