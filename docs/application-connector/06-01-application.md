---
title: Application
type: Custom Resource
---

The `applications.applicationconnector.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to register an Application in Kyma. The `Application` custom resource (CR) defines the APIs that the Application offers. After creating a new custom resource for a given Application, the Application is mapped to appropriate ServiceClasses in the Service Catalog. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd applications.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that registers the `system-prod` Application which offers one service.

>**NOTE:** The name of the Application must consist of lower case alphanumeric characters, `-` or `.`, and start and end with an alphanumeric character.

```yaml
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: Application
metadata:
  name: system-prod
spec:
  description: This is the system-production Application.
  labels:
    region: us
    kind: production
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **spec.description** | No | Describes the connected Application.  |
| **spec.accessLabel** | No | Labels the Application when an ApplicationMapping is created. |
| **spec.labels** | No | Defines the labels of the Application. |
| **spec.services** | No | Contains all services that the Application provides. |
| **spec.services.id** | Yes | Identifies the service that the Application provides. |
| **spec.services.identifier** | No | Provides an additional identifier of the ServiceClass. |
| **spec.services.name** | No | Represents a unique name of the service used by the Service Catalog. |
| **spec.services.displayName** | Yes | Specifies a human-readable name of the Application service. Parameter provided by the Application Registry, do not edit. |
| **spec.services.description** | No | Provides a short, human-readable description of the service offered by the Application. Parameter provided by the Application Registry, do not edit. |
| **spec.services.longDescription** | No | Provides a longer, human-readable description of the service offered by the Application. Parameter provided by the Application Registry, do not edit. |
| **spec.services.providerDisplayName** | Yes | Specifies a human-readable name of the Application service provider. Parameter provided by the Application Registry, do not edit. |
| **spec.services.tags** | No | Specifies additional tags used for better documentation of the available APIs. Parameter provided by the Application Registry, do not edit. |
| **spec.services.labels** | No | Specifies additional labels for the service offered by the Application. Parameter provided by the Application Registry, do not edit. |
| **spec.services.entries** | Yes | Contains the information about the APIs and events that the service offered by the Application provides. Parameter provided by the Application Registry, do not edit. |
| **spec.services.entries.type** | Yes | Specifies the entry type: API or event. Parameter provided by the Application Registry, do not edit. |
| **spec.services.entries.gatewayUrl** | No | Specifies the URL of the Application Connector. This field is required for the API entry type. Parameter provided by the Application Registry, do not edit. |
| **spec.services.entries.accessLabel** | No | Specifies the label used in Istio rules in the Application Connector. This field is required for the API entry type. |
| **spec.services.entries.targetUrl** |  No | Specifies the URL of a given API. This field is required for the API entry type. Parameter provided by the Application Registry, do not edit. |
| **spec.services.entries.oauthUrl** | No | Specifies the URL used to authorize with a given API. This field is required for the API entry type. Parameter provided by the Application Registry, do not edit. |
| **spec.services.entries.credentialsSecretName** | No | Specifies the name of the Secret which allows you to call a given API. This field is required if **spec.services.entries.oauthUrl** is specified. Parameter provided by the Application Registry, do not edit. |

## Related resources and components

These components use this CR:

| Component   |  Description |
|-----------|-------------|
| Application Registry | Reads and saves the APIs and Event Catalog metadata of the connected external solution in this CR. |
| Application Broker | Exposes the APIs and event definitions stored in this CR as ServiceClasses to the Service Catalog. |
| Application Operator | Provisions and de-provisions an instance of Application Gateway and Event Service for every created or deleted Application CR. |

## Additional information

The Application Operator adds the **status** section which describes the status of the Application installation to the created CR periodically. This table lists the fields of the **status** section.

| Field   |  Description |
|----------|-------------|
| **status.installationStatus** | Describes the status of the Application installation. |
| **status.installationStatus.description** | Provides a longer description of the installation status. |
| **status.installationStatus.status** | Provides a short, human-readable description of the installation status. |
