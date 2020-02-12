---
title: CompassConnection
type: Custom Resource
---

The `compassconnections.compass.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to prevent the status of the connection between Runtime Agent and Compass. The `CompassConnection` custom resource (CR) contains the connection statuses and Compass URLs. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd compassconnections.compass.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that registers the `compass-agent-connection` CompassConnection which preserves te status of the connection between Runtime Agent and Compass.

```yaml
apiVersion: compass.kyma-project.io/v1alpha1
kind: CompassConnection
metadata:
  name: compass-agent-connection
  connectorUrl: {CONNECTOR_URL}
  directorUrl: {DIRECTOR_URL}
  connectionStatus:
    certificateStatus: {DATE_CERT_ISSUED}
    established: {DATE_CONN_ESTABLISHED}
    lastSuccess: {LAST_SUCCESSFUL_SYNC_WITH_CONNECTOR}
    lastSync: {LAST_SYNC_ATTEMPT}
    renewed: {LAST_TIME_CERT_RENEWED}
  synchronizationStatus:
    lastAttempt: {LAST_SYNC_ATTEMPT_WITH_DIRECTOR}
    lastSuccessfulFetch: {LAST_SUCCESSFUL_RESOURCES_FETCH_FROM_DIRECTOR}
    lastSuccessfulApplication: {LAST_SUCCESSFUL_COMPASS_RESOURCES_APPLICATION}
{another_field}:
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| **Parameter** | **Required** | **Description** |
|---------------|:------------:|-----------------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **** |  |  |

## Related resources and components

These are the resources related to this CR:

| **Custom Resource** | **Description**  |
|---------------------|:----------------:|
|                     |                  |

These components use this CR:

| **Component** | **Description** |
|---------------|:---------------:|
|               |                 |

## Additional information

<!--- 
The Application Operator adds the **status** section which describes the status of the Application installation to the created CR periodically. This table lists the fields of the **status** section.

| Field   |  Description |
|----------|-------------|
| **status.installationStatus** | Describes the status of the Application installation. |
| **status.installationStatus.description** | Provides a longer description of the installation status. |
| **status.installationStatus.status** | Provides a short, human-readable description of the installation status. |
--->