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
  clusterName: ""
  creationTimestamp: 2018-09-25T12:24:41Z
  generation: 1
  labels:
    app: ec-default-gateway
    heritage: Tiller-gateway
    release: ec-default-gateway
  name: ec-default
  namespace: ""
  resourceVersion: "58761"
  selfLink: /apis/applicationconnector.kyma-project.io/v1alpha1/remoteenvironments/ec-default
  uid: fa2d9272-c0bd-11e8-ad1b-000d3a2fa0d4
spec:
  accessLabel: echo-access-label
  description: This Remote Environment corresponds to the connected remote system.
  labels: null
  services:
  - description: Events v1
    displayName: Events v1
    entries:
    - credentialsSecretName: ""
      gatewayUrl: ""
      oauthUrl: ""
      targetUrl: ""
      type: Events
    id: 40d3f17a-b376-4b02-8755-f24ceb76b27a
    identifier: ec-all-events
    labels:
      connected-app: ec-default
    longDescription: Events
    name: ec-events-v1-bbd7f
    providerDisplayName: Provider
  - description: Commerce Webservices
    displayName: Commerce Webservices
    entries:
    - accessLabel: re-ec-default-58cc45cd-b9ca-4c53-8019-0296774b7aa1
      credentialsSecretName: re-ec-default-58cc45cd-b9ca-4c53-8019-0296774b7aa1
      gatewayUrl: http://re-ec-default-58cc45cd-b9ca-4c53-8019-0296774b7aa1.kyma-integration.svc.cluster.local
      oauthUrl: https://oauth/token
      targetUrl: https://rest/v2
      type: API
    id: 58cc45cd-b9ca-4c53-8019-0296774b7aa1
    identifier: ""
    labels:
      connected-app: ec-default
    longDescription: EC OCC Commerce Webservices v2
    name: ec-occ-commerce-webservices-v2-a4508
    providerDisplayName: Provider
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **spec.source** |    **NO**   | Identifies the Remote Environment in the cluster. |
| **spec.description** |    **NO**   | Describes the connected Remote Environment.  |
| **spec.accessLabel** |    **NO**   | Labels the environment when the [EnvironmentMapping](041-cr-environment-mapping.md) is created. |
| **spec.labels** |    **NO**   | Labels the identified Remote Environment's taxonomy. |
| **spec.services** |    **NO**   | Contains all services that the Remote Environment provides. |
| **spec.services.id** |    **YES**   | Identifies the service that the Remote Environment provides. |
| **spec.services.identifier** |    **NO**   | Provides an additional identifier of the ServiceClass. |
| **spec.services.name** |    **NO**   | Represents a unique name of the service used by the Service Catalog. |
| **spec.services.displayName** |    **YES**   | Specifies a human-readable name of the Remote Environment service. |
| **spec.services.description** |    **NO**   | Provides a short, human-readable description of the Remote Environment service. |
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