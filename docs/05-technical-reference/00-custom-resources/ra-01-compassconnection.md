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
### CompassConnection.compass.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **managementInfo** (required) | object |  |
| **managementInfo.&#x200b;connectorUrl** (required) | string | Connector URL used for maintaining secure connection. |
| **managementInfo.&#x200b;directorUrl** (required) | string | Director URL used for fetching Applications |
| **refreshCredentialsNow**  | boolean | If true - ignore certificate expiration date and refresh next round |
| **resyncNow**  | boolean | If true - ignore `APP_MINIMAL_COMPASS_SYNC_TIME` and sync next round |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **connectionState** (required) | string |  |
| **connectionStatus** (required) | object | ConnectionStatus represents status of a connection to Compass |
| **connectionStatus.&#x200b;certificateStatus** (required) | object | Provides the certificate issue and expiration dates. |
| **connectionStatus.&#x200b;certificateStatus.&#x200b;acquired**  | string | When the certificate was acquired |
| **connectionStatus.&#x200b;certificateStatus.&#x200b;notAfter**  | string | When the certificate stops being valid |
| **connectionStatus.&#x200b;certificateStatus.&#x200b;notBefore**  | string | When the certificate becomes valid |
| **connectionStatus.&#x200b;error**  | string |  |
| **connectionStatus.&#x200b;established**  | string | Provides the date of when the connection was established |
| **connectionStatus.&#x200b;lastSuccess**  | string | Provides the date of the last successful synchronization with the Connector |
| **connectionStatus.&#x200b;lastSync**  | string | Provides the date of the last synchronization attempt |
| **connectionStatus.&#x200b;renewed**  | string | Provides the date of the last certificate renewal |
| **synchronizationStatus**  | object | Describes the status of the synchronization with the Director |
| **synchronizationStatus.&#x200b;error**  | string |  |
| **synchronizationStatus.&#x200b;lastAttempt**  | string | Provides the date of the last synchronization attempt with the Director |
| **synchronizationStatus.&#x200b;lastSuccessfulApplication**  | string | Provides the date of the last successful application of resources fetched from Compass |
| **synchronizationStatus.&#x200b;lastSuccessfulFetch**  | string | Provides the date of the last successful fetch of resources from the Director |

<!-- TABLE-END -->

## Dependents

| **Component** | **Description**                                                                                            |
|---------------|------------------------------------------------------------------------------------------------------------|
| Runtime Agent | Stores the Connector and Director URLs and preserves the status of the connection with Compass in this CR. |

