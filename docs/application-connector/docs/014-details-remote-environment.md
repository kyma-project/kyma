---
title: RemoteEnvironment custom resource
type: Details
---

## Overview

This file contains information about the RemoteEnvironment custom resource.
The RemoteEnvironment resource registers a remote environment in Kyma. The RemoteEnvironment resource defines APIs that the remote environment offers, such as Orders API in the EC. As a result, the RemoteEnvironment is mapped to service classes in the Service Catalog.

## Description

The RemoteEnvironment **spec** field contains the following attributes:
 * **source** which identifies the remote environment in the cluster
 * **services** which contains all services that the remote environment provides
 * **accessLabel** which labels the environment (Kubernetes Namespace)

The RemoteEnvironment **spec.services** list contains objects with the following fields:
 * **id** is a required unique field that the UI uses to fetch JSON schemas and documents. This filed maps to the **metadata.remoteEnvironmentServiceId** OSB service attribute.
 * **displayName** is a required field which maps to the **metadata.displayName** OSB service attribute. It is normalized and mapped to the **name** field from the [OSB Service object specification](https://github.com/openservicebrokerapi/servicebroker/blob/v2.12/spec.md#service-objects).
 * **longDescription** is a required field which maps to the **metadata.longDescription** OSB service attribute.
 * **providerDisplayName** is a required field which maps to the **metadata.providerDisplayName** OSB service attribute.
 * **tags** is an optional field which maps to the **tags** OSB service attribute. Tags provide a flexible mechanism to expose a classification, attribute, or base technology of a service.
 * **entries** is a field that contains information about APIs and events. This is a collection which must contain at least one element,
   and at most one element of the API type and one of the event type. An API entry must contain **gatewayUrl** and **accessLabel** fields. **accessLabel** must be unique for all the services of all RemoteEnvironments.

The OSB Service contains one default plan.

## Example

This is an example of the RemoteEnvironment custom resource:

```yaml
apiVersion: remoteenvironment.kyma.cx/v1alpha1
kind: RemoteEnvironment
metadata:
  name: re-prod
spec:
  source:
    environment: "production"
    type: "commerce"
    namespace: "com.github"
  description: "RE description"
  accessLabel: "re-access-label"
  services:
    - id: "ac031e8c-9aa4-4cb7-8999-0d358726ffaa"

      displayName: "Promotions"
      longDescription: "Promotions APIs"
      providerDisplayName: "Organization name"

      tags:
      - occ
      - Promotions

      entries:
      - type: API
        gatewayUrl: "http://promotions-gateway.production.svc.cluster.local"
        accessLabel: "access-label-1"
        targetUrl: "http://10.0.0.54:9932/occ/promotions"
        oauthUrl: "http://10.0.0.55:10219/occ/token"
        credentialsSecretName: "re-ac031e8c-9aa4-4cb7-8999-0d358726ffaa"
      - type: Events
```

This Remote Environment is mapped to the OSB Service:

```json
{
  "name":        "ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
  "id":          "ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
  "description": "Promotions APIs",
  "metadata": {
    "displayName":         "Promotions",
    "longDescription":     "Promotions APIs",
    "providerDisplayName": "Organization name",

    "labelsRequiredOnInstance": "access-label-1",
    "remoteEnvironmentServiceId": "ac031e8c-9aa4-4cb7-8999-0d358726ffaa",
    "source": {
          "environment": "production",
          "type": "commerce",
          "namespace": "com.github"
     }
  },
  "tags": ["occ", "promotions"],

  "plans":[
    {
      "name": "default",
      "id": "global unique GUID",
      "description": "Default plan",
      "metadata": {
        "displayName": "Default"
      }
    }
  ]
}
```
