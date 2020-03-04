---
title: HTTPSource
type: Custom Resource
---

The `httpsources.sources.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to define an event source in Kyma.
The HTTP Source custom resource (CR) defines the source of events. This means it specifies the Application that sends these events.
To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd httpsources.sources.kyma-project.io -o yaml
```

## Sample custom resource

This is a sample resource that receives events from an Application.

```yaml
apiVersion: sources.kyma-project.io/v1alpha1
kind: HTTPSource
metadata:
  name: varkes
  namespace: prod
spec:
  source: {application_name}
```
##  Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **metadata.namespace** | Yes | Specifies the Namespace in which the CR is created. |
| **spec.source** | Yes | Specifies a human-readable name of the Application that sends the events. |
