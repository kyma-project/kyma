---
title: RemoteEnvironment
type: Custom Resource
---

The `remoteenvironments.applicationconnector.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to register a Remote Environment (RE) in Kyma. The `RemoteEnvironment` custom resource defines the APIs that the RE offers. After creating a new custom resource for a given RE, the RE is mapped to appropriate ServiceClasses in the Service Catalog. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd remoteenvironments.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that registers the `re-prod` Remote Environment which offers one service.

```
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: RemoteEnvironment
metadata:
  name: system_prod
spec:
  description: This is the system_production Remote Environment.
  labels:
    region: us
    kind: production
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **spec.source** |    **NO**   | Identifies the Remote Environment in the cluster. |
| **spec.description** |    **NO**   | Describes the connected Remote Environment.  |
| **spec.accessLabel** |    **NO**   | Labels the RE when an EnvironmentMapping is created. |
| **spec.labels** |    **NO**   | Defines the labels of the RE. |
| **spec.services** |    **NO**   | Contains all services that the Remote Environment provides. |
| **spec.services.id** |    **YES**   | Identifies the service that the Remote Environment provides. |
| **spec.services.identifier** |    **NO**   | Provides an additional identifier of the ServiceClass. |
| **spec.services.name** |    **NO**   | Represents a unique name of the service used by the Service Catalog. |
| **spec.services.displayName** |    **YES**   | Specifies a human-readable name of the Remote Environment service. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.description** |    **NO**   | Provides a short, human-readable description of the service offered by the RE. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.longDescription** |    **NO**   | Provides a longer, human-readable description of the service offered by the RE. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.providerDisplayName** |    **YES**   | Specifies a human-readable name of the Remote Environment service provider. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.tags** |    **NO**   | Specifies additional tags used for better documentation of the available APIs. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.labels** |    **NO**   | Specifies additional labels for the service offered by the RE. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.entries** |    **YES**   | Contains the information about the APIs and Events that the service offered by the RE provides. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.entries.type** |    **YES**   | Specifies the entry type: API or Event. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.entries.gatewayUrl** |    **NO**   | Specifies the URL of the Application Connector. This field is required for the API entry type. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.entries.accessLabel** |    **NO**   | Specifies the label used in Istio rules in the Application Connector. This field is required for the API entry type. |
| **spec.services.entries.targetUrl** |    **NO**   | Specifies the URL of a given API. This field is required for the API entry type. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.entries.oauthUrl** |    **NO**   | Specifies the URL used to authorize with a given API. This field is required for the API entry type. Parameter provided by the Metadata Service, do not edit. |
| **spec.services.entries.credentialsSecretName** |    **NO**   | Specifies the name of the Secret which allows you to call a given API. This field is required if **spec.services.entries.oauthUrl** is specified. Parameter provided by the Metadata Service, do not edit. |

## Additional information

The Application Operator adds the **status** section which describes the status of the RE installation to the created CR periodically. This table lists the fields of the **status** section.

| Field   |  Description |
|:----------:|:-------------:|
| **status.installationStatus** | Describes the status of the RE installation. |
| **status.installationStatus.description** | Provides a longer description of the installation status. |
| **status.installationStatus.status** | Provides a short, human-readable description of the installation status. |
