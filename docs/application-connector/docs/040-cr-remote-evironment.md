---
title: RemoteEnvironment
type: Custom Resource
---

The `remoteenvironments.applicationconnector.kyma-project.io` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format used to register a Remote Environment in Kyma. The RemoteEnvironment resource defines APIs that the Remote Environment offers. As a result, the RemoteEnvironment is mapped to ServiceClasses in the Service Catalog. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd remoteenvironments.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that registers the `re-prod` Remote Environment which provides one service with the `ac031e8c-9aa4-4cb7-8999-0d358726ffaa` ID.

```
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: RemoteEnvironment
metadata:
  name: re-prod
spec:
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

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **spec.source** |    **NO**   | Identifies the Remote Environment in the cluster. |
| **spec.description** |    **NO**   | Describes the connected Remote Environment.  |
| **spec.accessLabel** |    **NO**   | Labels the environment when the [EnvironmentMapping](041-cr-environment-mapping.md) is created. |
| **spec.labels** |    **NO**   | Labels indentified Remote Environment taxonomy.  |
| **spec.services** |    **NO**   | Contains all services that the Remote Environment provides. |
| **spec.services.id** |    **YES**   | Identifies the service that the Remote Environment provides. |
| **spec.services.identifier** |    **NO**   | Additional identifier of the service class. |
| **spec.services.name** |    **NO**   | Unique name of the service used by the Service Catalog. |
| **spec.services.displayName** |    **YES**   | Specifies a human-readable name of the Remote Environment service. |
| **spec.services.description** |    **NO**   | Provides a short human-readable description of the Remote Environment service. |
| **spec.services.longDescription** |    **NO**   | Provides a human-readable description of the Remote Environment service. |
| **spec.services.providerDisplayName** |    **YES**   | Specifies a human-readable name of the Remote Environment service provider. |
| **spec.services.tags** |    **NO**   | Specifies the categories of the Remote Environment service. |
| **spec.services.labels** |    **NO**   | Specifies the additional labels of the Remote Environment service. |
| **spec.services.entries** |    **YES**   | Contains information about APIs and Events that the Remote Environment service provides. |
| **spec.services.entries.type** |    **YES**   | Specifies whether the entry is of API or Event type. |
| **spec.services.entries.gatewayUrl** |    **NO**   | Specifies the URL of the Application Connector. This field is required for the API entry type. |
| **spec.services.entries.accessLabel** |    **NO**   | Specifies the label used in Istio rules in the Application Connector. This field is required for the API entry type. |
| **spec.services.entries.targetUrl** |    **NO**   | Specifies the URL to a given API. This field is required for the API entry type. |
| **spec.services.entries.oauthUrl** |    **NO**   | Specifies the URL used to authorize with a given API. This field is required for the API entry type. |
| **spec.services.entries.credentialsSecretName** |    **NO**   | Specifies the name of the Secret which allows you to make a call to a given API. This field is required if the **spec.services.entries.oauthUrl** is specified. |
