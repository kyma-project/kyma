---
title: EventActivation
type: Custom Resource
---

The `eventactivations.remoteenvironment.kyma.cx` Custom Resource Definition (CRD) is a detailed description of the kind of data and the format used to create an Event Bus Subscription. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```
kubectl get crd eventactivations.remoteenvironment.kyma.cx -o yaml
```

## Sample custom resource

This is a sample resource that allows you to consume Events sent from the service with the `ac031e8c-9aa4-4cb7-8999-0d358726ffaa` ID in a `production` Namespace.

```
apiVersion: remoteenvironment.kyma.cx/v1alpha1
kind: EventActivation
metadata:
  name: "ac031e8c-9aa4-4cb7-8999-0d358726ffaa"
  namespace: production
spec:
  displayName: "Orders"
  source:
    environment: "prod"
    type: "commerce"
    namespace: "com.hakuna.matata"
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:


| Parameter   |      Mandatory?      |  Description |
|:----------:|:-------------:|:------|
| **metadata.name** |    **YES**   | Specifies the name of the CR and the Remote Environment service ID. |
| **metadata.namespace** |    **YES**   | Specifies the Namespace in which the CR is created. |
| **spec.displayName** |    **YES**   | Specifies a human-readable name of the Remote Environment service. |
| **spec.source** |    **YES**   | Identifies a Remote Environment in the cluster. |
| **spec.source.environment** |    **YES**   | Specifies the environment of the connected Remote Environment. |
| **spec.source.type** |    **YES**   | Specifies the type of the connected Remote Environment. |
| **spec.source.namespace** |    **YES**   | Specifies the namespace of the connected Remote Environment. |
