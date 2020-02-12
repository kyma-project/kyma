---
title: CompassConnection
type: Custom Resource
---

The `compassconnections.compass.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to prevent the status of the connection between Runtime Agent and Compass. The `CompassConnection` custom resource (CR) contains the connection statuses and Compass URLs. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd compassconnections.compass.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that registers the `compass-agent-connection` CompassConnection which preserves te status of the connection between Runtime Agent and Compass. It also stores the URLs for the Connector and the Director. 

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