---
title: EnvironmentMapping
type: Custom Resource
---

The `environmentmappings.remoteenvironment.kyma.cx` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format used to enable the Remote Environment with the name corresponding to the EnvironmentMapping in a given Namespace. To get the up-to-date CRD and show the output in the yaml format, run this command:

```
kubectl get crd environmentmappings.remoteenvironment.kyma.cx -o yaml
```

## Sample custom resource

This is a sample resource in which the EnvironmentMapping enables the `ec-prod` Remote Environment in the `production` Namespace:

```
apiVersion: remoteenvironment.kyma.cx/v1alpha1
kind: EnvironmentMapping
metadata:
  name: ec-prod
  namespace: production
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR and Remote Environment. |
| **metadata.namespace** |    **YES**   | Specifies the Namespace in which the Remote Environment is enabled. |
