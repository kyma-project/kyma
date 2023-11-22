---
title: Application CR
---

The `applications.applicationconnector.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to register an Application in Kyma. The `Application` custom resource (CR) defines the APIs that the Application offers. To get the up-to-date CRD and show the output in the YAML format, run this command:

```bash
kubectl get crd applications.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that registers the `system-prod` Application which offers one service.

>**NOTE:** The name of the Application must consist of lower case alphanumeric characters, `-` or `.`, and start and end with an alphanumeric character.

>**NOTE:** In case the Application name contains `-` or `.`, the underlying Eventing services uses a clean name with alphanumeric characters only. (For example, `system-prod` becomes `systemprod`).
>
> This could lead to a naming collision. For example, both `system-prod` and `systemprod` become `systemprod`.
>
> A solution for this is to provide an `application-type` label (with alphanumeric characters only) which is then used by the Eventing services instead of the Application name. If the `application-type` label also contains `-` or `.`, the underlying Eventing services clean it and use the cleaned label.

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
| **spec.skipVerify** | No | Determines whether to skip TLS certificate verification for the Application.  |
| **spec.encodeUrl** | No | Allows for URL encoding. If set to 'false', your URL segments stay intact. |
| **spec.labels** | No | Defines the labels of the Application. |
| **spec.services** | No | Contains all services that the Application provides. |
| **spec.services.id** | Yes | Identifies the service that the Application provides. |
| **spec.services.identifier** | No | Represents an additional identifier unique in the Application scope. Allows the external system to provide its own identifier. |
| **spec.services.name** | No | Represents a unique name of the service. Used for proxying in Application Gateway. |
| **spec.services.providerDisplayName** | Yes | Specifies a human-readable name of the Application service provider. In the [Compass mode](../../01-overview/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/application-connectivity/README.md), you have to provide it yourself. |
| **spec.services.tags** | No | Specifies additional tags used for better documentation of the available APIs. In the [Compass mode](../../01-overview/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/application-connectivity/README.md), you have to provide it yourself. |
| **spec.services.labels** | No | Specifies additional labels for the service offered by the Application. In the [Compass mode](../../01-overview/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/application-connectivity/README.md), you have to provide it yourself. |
| **spec.services.entries** | Yes | Contains the information about the APIs and events that the service offered by the Application provides. In the [Compass mode](../../01-overview/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/application-connectivity/README.md), you have to provide it yourself. |
| **spec.services.entries.type** | Yes | Specifies the entry type: `API` or `Events`. In the [Compass mode](../../01-overview/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/application-connectivity/README.md), you have to provide it yourself. |
| **spec.services.entries.centralGatewayUrl** | No | Specifies the URL of Application Gateway. Internal address resolvable only within the cluster. This field is required for the API entry type. In the [Compass mode](../../01-overview/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/application-connectivity/README.md), you have to provide it yourself. |
| **spec.services.entries.accessLabel** | No | Specifies the label used in Istio rules in Application Connector. This field is required for the API entry type. |
| **spec.services.entries.targetUrl** |  No | Specifies the URL of a given API. This field is required for the API entry type. In the [Compass mode](../../01-overview/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/application-connectivity/README.md), you have to provide it yourself. |
| **spec.services.entries.oauthUrl** | No | Specifies the URL used to authorize with a given API. This field is required for the API entry type. In the [Compass mode](../../01-overview/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/application-connectivity/README.md), you have to provide it yourself. |
| **spec.services.entries.credentialsSecretName** | No | Specifies the name of the Secret which allows you to call a given API. This field is required if **spec.services.entries.oauthUrl** is specified. In the [Compass mode](../../01-overview/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/application-connectivity/README.md), you have to provide it yourself. |

## Related resources and components

These components use this CR:

| Component   |  Description |
|-----------|-------------|
| Application Gateway | Reads the API metadata in order to connect to the external system. | 
| Application Connectivity Validator (in the [Compass mode](../../01-overview/application-connectivity/README.md)) | Validates requests and events from the external system against the respective Application CR. |
| Runtime Agent (in the [Compass mode](../../01-overview/application-connectivity/README.md)) | Saves the metadata of the connected external system in the Application CR, synchronizes the metadata stored in Compass with the state in the cluster stored in the Application CR. |