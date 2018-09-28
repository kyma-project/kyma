---
title: EnvironmentMapping
type: Custom Resource
---

The `environmentmappings.applicationconnector.kyma-project.io` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format used to enable APIs and Events from a Remote Environment as a ServiceClass in a given Namespace. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd environmentmappings.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource in which the EnvironmentMapping enables the `ec-prod` Remote Environment in the `production` Namespace:

```
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: EnvironmentMapping
metadata:
  name: ec-prod
  namespace: production
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR and the Remote Environment. |
| **metadata.namespace** |    **YES**   | Specifies the Namespace in which the Remote Environment is enabled. |
