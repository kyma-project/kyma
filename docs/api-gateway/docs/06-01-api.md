---
title: Api
type: Custom Resource
---

The `api.gateway.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the API Controller listens for. To get the up-to-date CRD and show
the output in the `yaml` format, run this command:
```
kubectl get crd apis.gateway.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample custom resource (CR) that the API-Controller listens for to expose a service. This example has the **authentication** section specified which makes the API Controller create an Istio Authentication Policy for this service.

```
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
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

| Field   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the exposed API |
| **service.name**, **service.port** | **YES** | Specifies the name and the communication port of the exposed service. |
| **spec.hostname** | **YES** | Specifies the service's external inbound communication address. |
| **spec.authentication** | **NO** | Allows to specify an array of authentication policies that secure the service. |
| **authentication.type** | **YES** | Specifies the desired authentication method that secures the exposed service. |
| **authentication.jwt.issuer**, **authentication.jwt.jwksUri** | **YES** | Specifies the issuer of the tokens used to access the services, as well as the JWKS endpoint URI. |


## Additional information

When you fetch an existing Api CR, the system adds the **status** section which describes the status of the Virtual Service and the Authentication Policy created for this CR. This table lists the fields of the **status** section.

| Field   |  Description |
|:----------:|:-------------:|
| **status.virtualService.code** | Status code describing the Virtual Service. |
| **status.virtualService.lastError** | Last error reported when creating or updating the Virtual Service. |
| **status.virtualService.resource.name** | Name of the created Virtual Service. |
| **status.virtualService.resource.uid** | UID of the created Virtual Service. |
| **status.virtualService.resource.version** | Version of the created Virtual Service. |
| **status.authenticationStatus.code** | Status code describing the Authentication Policy. |
| **status.authenticationStatus.lastError** | Last error reported when creating or updating the Authentication Policy. |
| **status.authenticationStatus.resource.name** | Name of the created Authentication Policy. |
| **status.authenticationStatus.resource.uid** | UID of the created Authentication Policy. |
| **status.authenticationStatus.resource.version** | Version of the created Authentication Policy. |

>**NOTE:** The **authenticationStatus** section is optional. It is created only when authentication is enabled for the given API.


### Status codes

These are the status codes used to describe the Authentication Policies and Virtual Services:

| Code   |  Description |
|:----------:|:-------------:|
| **0** | Resource not created. |
| **1** | Resource creation in progress. |
| **2** | Resource created successfully. |
| **3** | Error - resource not created. |
| **4** | Virtual Service-specific. Hostname already in use by a different Virtual Service. |
