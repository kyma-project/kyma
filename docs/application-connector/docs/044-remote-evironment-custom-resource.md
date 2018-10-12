---
title: RemoteEnvironment
type: Custom Resource
---

The `remoteenvironments.applicationconnector.kyma-project.io` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format used to register a Remote Environment (RE) in Kyma. The `RemoteEnvironment` Custom Resource defines the APIs that the RE offers. After creating a new Custom Resource for a given RE, the RE is mapped to appropriate ServiceClasses in the Service Catalog. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd remoteenvironments.applicationconnector.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that registers the `re-prod` Remote Environment which offers one service.

```
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: RemoteEnvironment
metadata:
  name: system_prod
spec:
  description: This is the system_production Remote Environment.
  labels:
    region: us
    kind: production
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR. |
| **spec.description** |    **NO**   | Provides a short, human-readable description of the RE. |
| **spec.labels** |    **NO**   | Defines the labels of the RE. |
