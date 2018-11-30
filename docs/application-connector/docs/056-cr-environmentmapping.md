---
title: EnvironmentMapping
type: Custom Resource
---

The `environmentmappings.remoteenvironment.kyma.cx` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to enable APIs and Events from a Remote Environment as a ServiceClass in a given Namespace. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd environmentmappings.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource in which the EnvironmentMapping enables the `test` Remote Environment in the `production` Environment:

```
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: EnvironmentMapping
metadata:
  name: test
  namespace: production
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR and the Remote Environment. |
| **metadata.namespace** |    **YES**   | Specifies the Namespace in which the Remote Environment is enabled. |

## Related resources and components

These are the resources related to this CR:

| Custom resource   |   Description |
|:----------:|:------|
| RemoteEnvironment |  Uses this CR to expose RemoteEnvironment's services in a given Environment. |

These components use this CR:

| Component   |   Description |
|:----------:|:------|
| Remote Environment Broker |  Uses this CR to enable the provisioning of ServiceClasses in a given Environment. |
| UI API Layer | Uses this CR to filter enabled RemoteEnvironments. It also allows you to create or delete EnvironmentMappings. |
