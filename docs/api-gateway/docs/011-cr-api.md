---
title: Api
type: Custom Resource
---

The `api.gateway.kyma-project.io` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format the API Controller listens for. To get the up-to-date CRD and show
the output in the `yaml` format, run this command:
```
kubectl get crd apis.gateway.kyma-project.io -o yaml
```

## Sample Custom Resource

This is a sample CR that the API-Controller listens for to expose a service. This example has the **authentication** section specified which makes the API Controller create an Istio Authentication Policy for this service.

```
apiVersion: gateway.kyma-project.io/v1alpha2
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


| Field   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the exposed API |
| **service.name**, **service.port** | **YES** | Specifies the name and the communication port of the exposed service. |
| **spec.hostname** | **YES** | Specifies the service's external inbound communication address. |
| **spec.authentication** | **NO** | Allows to specify an array of authentication policies that secure the service. |
| **authentication.type** | **YES** | Specifies the desired authentication method that secures the exposed service. |
| **authentication.jwt.issuer**, **authentication.jwt.jwksUri** | **YES** | Specifies the issuer of the tokens used to access the services, as well as the JWKS endpoint URI. |

Also when existing API is retrieved it contains additional **status** field. Status field describes status of two resources maintained by API: VirtualService and Policy. This table lists all possible status parameters:

| Field   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **status.virtualService** | **YES** | Section with statuses of related VirtualService |
| **status.virtualService.code** | **YES** | Status code of related VirtualService. See section **Status code** below |
| **status.virtualService.lastError** | **NO** | Last error reported while creating/updating VirtualService |
| **status.virtualService.resource** | **NO** | Section with information about created VirtualService. Not present if resource couldn't be created |
| **status.virtualService.resource.name** | **NO** | Name of created VirtualService |
| **status.virtualService.resource.version** | **NO** | Version of created VirtualService |
| **status.virtualService.resource.uid** | **NO** | Version of created VirtualService |
| **status.authenticationStatus** | **NO** | Section with statuses of related Policy |
| **status.authenticationStatus.code** | **NO** | Status code of related Policy. See section **Status code** below |
| **status.authenticationStatus.lastError** | **NO** | Last error reported while creating/updating Policy |
| **status.authenticationStatus.resource** | **NO** | Section with information about created Policy. Not present if resource couldn't be created |
| **status.authenticationStatus.resource.name** | **NO** | Name of created Policy |
| **status.authenticationStatus.resource.version** | **NO** | Version of created Policy |
| **status.authenticationStatus.resource.uid** | **NO** | Version of created Policy |

Section **authenticationStatus** is optional and will not be created if authentication is not enabled. 

### Status code

Following status codes are possible:

 0 - resource is not created
 1 - resource creation is in progress
 2 - resource is created
 3 - there was an error during resource creation