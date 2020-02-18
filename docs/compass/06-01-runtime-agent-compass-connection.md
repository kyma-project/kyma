---
title: CompassConnection
type: Custom Resource
---

The `compassconnections.compass.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to preserve the status of the connection between Runtime Agent and Compass. The `CompassConnection` custom resource (CR) contains the connection statuses and Compass URLs. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd compassconnections.compass.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that registers the `compass-agent-connection` CompassConnection which preserves the status of the connection between Runtime Agent and Compass. It also stores the URLs for the Connector and the Director. 

```yaml
apiVersion: compass.kyma-project.io/v1alpha1
kind: CompassConnection
metadata:
  name: compass-connection
spec:
  managementInfo:
    connectorUrl: https://compass-gateway-mtls.34.76.33.244.xip.io/connector/graphql
    directorUrl: https://compass-gateway-mtls.34.76.33.244.xip.io/director/graphql
status:
  connectionState: ConnectionMaintenanceFailed
  connectionStatus:
    certificateStatus:
      acquired: "2020-02-11T10:35:22Z"
      notAfter: "2020-05-11T10:35:22Z"
      notBefore: "2020-02-11T10:35:22Z"
    established: "2020-02-11T10:35:22Z"
    lastSuccess: "2020-02-12T10:45:10Z"
    lastSync: "2020-02-12T12:37:48Z"
    renewed: null
  synchronizationStatus:
    lastAttempt: "2020-02-12T10:45:10Z"
    lastSuccessfulApplication: "2020-02-12T10:45:10Z"
    lastSuccessfulFetch: "2020-02-12T10:45:10Z"
```

## Custom resource parameters

This table lists all the possible parameters of the CompassConnection custom resource together with their descriptions:

| **Parameter** | **Required** | **Description** |
|---------------|:------------:|-----------------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **spec.managementInfo.connectorUrl** | Yes | Connector URL used for maintaining secure connection. |
| **spec.managementInfo.directorUrl** | Yes | Director URL used for fetching Applications. |

These components use this CR:

| **Component** | **Description** |
|---------------|-----------------|
| Runtime Agent | Stores the Connector and Director URLs and preserves the status of the connection with Compass in this CR. |

## Additional information

Runtime Agent adds the **status** section which describes the statuses of the connection and synchronization to the created CR periodically. This table lists the fields of the **status** section.

| Field   |  Description |
|---------|-------------|
| **status.connectionStatus** | Describes the status of the connection with Compass. |
| **status.connectionStatus.certificateStatus** | Provides the dates of when the certificate was issued and when it expires. |
| **status.connectionStatus.established** | Provides the date of when the connection was established. |
| **status.connectionStatus.lastSuccess** | Provides the date of the last successful synchronization with the Connector. |
| **status.connectionStatus.lastSync** | Provides the date of the last synchronization attempt. |
| **status.connectionStatus.renewed** | Provides the date of the last certificate renewal. |
| **status.synchronizationStatus** | Describes the status of the synchronization with the Director. |
| **status.synchronizationStatus.lastAttempt** | Provides the date of the last synchronization attempt with the Director. |
| **status.synchronizationStatus.lastSuccessfulFetch** | Provides the date of the last successful fetch of resources from the Director. |
| **status.synchronizationStatus.lastSuccessfulApplication** | Provides the date of the last successful application of resources fetched from Compass. |