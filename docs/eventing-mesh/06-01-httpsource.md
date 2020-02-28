---
title: HTTPSource
type: Custom Resource
---

The `httpsources.sources.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to create and event source adapter in Kyma.
The HTTP Source custom resource (CR) defines the sources of events accepted by the HTTP server.To get the up-to-date CRD and show the output in the yaml format, run this command:

```bash
kubectl get crd httpsources.sources.kyma-project.io -o yaml
```

## Sample custom resource

## 

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **metadata.namespace** | Yes | Specifies the Namespace in which the CR is created. |
| **spec.displayName** | Yes | Specifies a human-readable name of the Application service. |
| **spec.sourceId** | Yes | Used to construct a Publish-Subscribe (Pub/Sub) topic name where events are sent and from where they are consumed. |