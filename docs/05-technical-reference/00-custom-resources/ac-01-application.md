---
title: Application
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

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->
<!-- TABLE-START -->
<!-- Application v1alpha1 applicationconnector.kyma-project.io -->
| Parameter         | Description                                   |
| ---------------------------------------- | ---------|
| **spec.accessLabel** |  |
| **spec.compassMetadata** |  |
| **spec.compassMetadata.applicationId** |  |
| **spec.compassMetadata.authentication** |  |
| **spec.compassMetadata.authentication.clientIds** |  |
| **spec.description** | Describes the connected Application. |
| **spec.displayName** |  |
| **spec.encodeUrl** | Allows for URL encoding. If set to `false`, your URL segments stay intact. |
| **spec.group** |  |
| **spec.labels** | Defines the labels of the Application. |
| **spec.longDescription** |  |
| **spec.providerDisplayName** |  |
| **spec.services** | Contains all services that the Application provides. |
| **spec.services.authCreateParameterSchema** | New fields used by V2 version |
| **spec.services.description** |  |
| **spec.services.displayName** |  |
| **spec.services.entries** | Contains the information about the APIs and events that the service offered by the Application provides. In the [Compass mode](../../01-overview/main-areas/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/main-areas/application-connectivity/README.md), you have to provide it yourself.
 |
| **spec.services.entries.accessLabel** | Specifies the label used in Istio rules in Application Connector. This field is required for the API entry type.
 |
| **spec.services.entries.apiType** |  |
| **spec.services.entries.centralGatewayUrl** | Specifies the URL of Application Gateway. Internal address resolvable only within the cluster. This field is required for the API entry type. In the [Compass mode](../../01-overview/main-areas/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/main-areas/application-connectivity/README.md), you have to provide it yourself.
 |
| **spec.services.entries.credentials** |  |
| **spec.services.entries.credentials.authenticationUrl** |  |
| **spec.services.entries.credentials.csrfInfo** |  |
| **spec.services.entries.credentials.csrfInfo.tokenEndpointURL** |  |
| **spec.services.entries.credentials.secretName** |  |
| **spec.services.entries.credentials.type** |  |
| **spec.services.entries.gatewayUrl** |  |
| **spec.services.entries.id** |  |
| **spec.services.entries.name** | New fields used by V2 version |
| **spec.services.entries.requestParametersSecretName** |  |
| **spec.services.entries.specificationUrl** |  |
| **spec.services.entries.targetUrl** | Specifies the URL of a given API. This field is required for the API entry type. In the [Compass mode](../../01-overview/main-areas/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/main-areas/application-connectivity/README.md), you have to provide it yourself.
 |
| **spec.services.entries.type** | Specifies the entry type: `API` or `Events`. In the [Compass mode](../../01-overview/main-areas/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/main-areas/application-connectivity/README.md), you have to provide it yourself.
 |
| **spec.services.id** | Identifies the service that the Application provides. |
| **spec.services.identifier** | Represents an additional identifier unique in the Application scope. Allows the external system to provide its own identifier.
 |
| **spec.services.labels** | Deprecated |
| **spec.services.longDescription** |  |
| **spec.services.name** | Represents a unique name of the service. Used for proxying in Application Gateway.
 |
| **spec.services.providerDisplayName** | Specifies a human-readable name of the Application service provider. In the [Compass mode](../../01-overview/main-areas/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/main-areas/application-connectivity/README.md), you have to provide it yourself.
 |
| **spec.services.tags** | Specifies additional tags used for better documentation of the available APIs. In the [Compass mode](../../01-overview/main-areas/application-connectivity/README.md), it's provided by Runtime Agent. In the [standalone mode](../../01-overview/main-areas/application-connectivity/README.md), you have to provide it yourself.
 |
| **spec.skipInstallation** |  |
| **spec.skipVerify** | Determines whether to skip TLS certificate verification for the Application. |
| **spec.tags** | New fields used by V2 version. |
| **spec.tenant** |  |
| **status.installationStatus** | Represents the status of Application release installation |
| **status.installationStatus.description** |  |
| **status.installationStatus.status** |  |<!-- TABLE-END -->


## Related resources and components

These components use this CR:

| Component                                                                                                                   | Description                                                                                                                                                                        |
|-----------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Application Gateway                                                                                                         | Reads the API metadata in order to connect to the external system.                                                                                                                 |
| Application Connectivity Validator (in the [Compass mode](../../01-overview/main-areas/application-connectivity/README.md)) | Validates requests and events from the external system against the respective Application CR.                                                                                      |
| Runtime Agent (in the [Compass mode](../../01-overview/main-areas/application-connectivity/README.md))                      | Saves the metadata of the connected external system in the Application CR, synchronizes the metadata stored in Compass with the state in the cluster stored in the Application CR. |
