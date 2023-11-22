---
title: CompassConnection CR
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

This table lists all the possible parameters of the CompassConnection custom resource together with their descriptions. For more details, see the [CompassConnection specification file](../../../installation/resources/crds/compass-runtime-agent/compass-connection.crd.yaml).

<!-- The table below was generated automatically -->
<!-- Some special tags (html comments) are at the end of lines due to markdown requirements. -->
<!-- The content between "TABLE-START" and "TABLE-END" will be replaced -->

<!-- TABLE-START -->
### CompassConnection.compass.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **managementInfo** (required) | object |  |
| **managementInfo.&#x200b;connectorUrl** (required) | string | URL used for maintaining the secure connection. |
| **managementInfo.&#x200b;directorUrl** (required) | string | URL used for fetching Applications. |
| **refreshCredentialsNow**  | boolean | If set to `true`, ignores certificate expiration date and refreshes in the next round. |
| **resyncNow**  | boolean | If set to `true`, ignores `APP_MINIMAL_COMPASS_SYNC_TIME` and syncs in the next round. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **connectionState** (required) | string |  |
| **connectionStatus** (required) | object | Represents the status of the connection to Compass. |
| **connectionStatus.&#x200b;certificateStatus** (required) | object | Specifies the certificate issue and expiration dates. |
| **connectionStatus.&#x200b;certificateStatus.&#x200b;acquired**  | string | Specifies when the certificate was acquired. |
| **connectionStatus.&#x200b;certificateStatus.&#x200b;notAfter**  | string | Specifies when the certificate stops being valid. |
| **connectionStatus.&#x200b;certificateStatus.&#x200b;notBefore**  | string | Specifies when the certificate becomes valid. |
| **connectionStatus.&#x200b;error**  | string |  |
| **connectionStatus.&#x200b;established**  | string | Specifies when the connection was established. |
| **connectionStatus.&#x200b;lastSuccess**  | string | Specifies the date of the last successful synchronization with the Connector. |
| **connectionStatus.&#x200b;lastSync**  | string | Specifies the date of the last synchronization attempt. |
| **connectionStatus.&#x200b;renewed**  | string | Specifies the date of the last certificate renewal. |
| **synchronizationStatus**  | object | Provides the status of the synchronization with the Director. |
| **synchronizationStatus.&#x200b;error**  | string |  |
| **synchronizationStatus.&#x200b;lastAttempt**  | string | Specifies the date of the last synchronization attempt with the Director. |
| **synchronizationStatus.&#x200b;lastSuccessfulApplication**  | string | Specifies the date of the last successful application of resources fetched from Compass. |
| **synchronizationStatus.&#x200b;lastSuccessfulFetch**  | string | Specifies the date of the last successful fetch of resources from the Director. |

<!-- TABLE-END -->

## Dependents

| **Component** | **Description**                                                                                            |
|---------------|------------------------------------------------------------------------------------------------------------|
| Runtime Agent | Stores the Connector and Director URLs and preserves the status of the connection with Compass in this CR. |

