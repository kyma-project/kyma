---
title: Compass Connection
---

The `compassconnections.compass.kyma-project.io` CustomResourceDefinition (CRD) 
is a detailed description of the kind of data and the format used to preserve 
the status of the connection between the Runtime Agent and Compass. 
The `CompassConnection` custom resource (CR) contains the connection statuses and Compass URLs.
To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd compassconnections.compass.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that registers the `compass-agent-connection` CompassConnection
which preserves the status of the connection between Runtime Agent and Compass. 
It also stores the URLs for the Connector and the Director.

```yaml
apiVersion: compass.kyma-project.io/v1alpha1
kind: CompassConnection
metadata:
  name: compass-connection
spec:
  managementInfo:
    connectorUrl: https://compass-gateway-mtls.kyma.example.com/connector/graphql
    directorUrl: https://compass-gateway-mtls.kyma.example.com/director/graphql
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

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->

<!-- TABLE-START -->
<!-- CompassConnection v1alpha1 compass.kyma-project.io -->
| Parameter                                                  | Description                                                                            |
|------------------------------------------------------------|----------------------------------------------------------------------------------------|
| **spec.managementInfo**                                    |                                                                                        |
| **spec.managementInfo.connectorUrl**                       | Connector URL used for maintaining secure connection.                                  |
| **spec.managementInfo.directorUrl**                        | Director URL used for fetching Applications                                            |
| **spec.refreshCredentialsNow**                             | If true - ignore certificate expiration date and refresh next round                    |
| **spec.resyncNow**                                         | If true - ignore `APP_MINIMAL_COMPASS_SYNC_TIME` and sync next round                   |
| **status.connectionState**                                 |                                                                                        |
| **status.connectionStatus**                                | ConnectionStatus represents status of a connection to Compass                          |
| **status.connectionStatus.certificateStatus**              | Provides the dates of when the certificate was issued and when it expires.             |
| **status.connectionStatus.certificateStatus.acquired**     | When the certificate was acquired                                                      |
| **status.connectionStatus.certificateStatus.notAfter**     | When the certificate stops being valid                                                 |
| **status.connectionStatus.certificateStatus.notBefore**    | When the certificate becomes valid                                                     |
| **status.connectionStatus.error**                          |                                                                                        |
| **status.connectionStatus.established**                    | Provides the date of when the connection was established                               |
| **status.connectionStatus.lastSuccess**                    | Provides the date of the last successful synchronization with the Connector            |
| **status.connectionStatus.lastSync**                       | Provides the date of the last synchronization attempt                                  |
| **status.connectionStatus.renewed**                        | Provides the date of the last certificate renewal                                      |
| **status.synchronizationStatus**                           | Describes the status of the synchronization with the Director                          |
| **status.synchronizationStatus.error**                     |                                                                                        |
| **status.synchronizationStatus.lastAttempt**               | Provides the date of the last synchronization attempt with the Director                |
| **status.synchronizationStatus.lastSuccessfulApplication** | Provides the date of the last successful application of resources fetched from Compass |
| **status.synchronizationStatus.lastSuccessfulFetch**       | Provides the date of the last successful fetch of resources from the Director          |
<!-- TABLE-END -->

## Dependents

| **Component** | **Description**                                                                                            |
|---------------|------------------------------------------------------------------------------------------------------------|
| Runtime Agent | Stores the Connector and Director URLs and preserves the status of the connection with Compass in this CR. |

