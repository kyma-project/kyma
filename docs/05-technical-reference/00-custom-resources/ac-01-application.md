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
<!-- TABLE-START --><!-- TABLE-END -->


## Related resources and components

These components use this CR:

| Component                                                                                                                   | Description                                                                                                                                                                        |
|-----------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Application Gateway                                                                                                         | Reads the API metadata in order to connect to the external system.                                                                                                                 |
| Application Connectivity Validator (in the [Compass mode](../../01-overview/main-areas/application-connectivity/README.md)) | Validates requests and events from the external system against the respective Application CR.                                                                                      |
| Runtime Agent (in the [Compass mode](../../01-overview/main-areas/application-connectivity/README.md))                      | Saves the metadata of the connected external system in the Application CR, synchronizes the metadata stored in Compass with the state in the cluster stored in the Application CR. |
