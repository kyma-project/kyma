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

This is a sample resource that receives events from an Applications and sends them to a preconfigured sink.

```
apiVersion: sources.kyma-project.io/v1alpha1
kind: HTTPSource
metadata:
  creationTimestamp: "2020-02-28T01:30:40Z"
  generation: 1
  name: varkes
  namespace: kyma-integration
  resourceVersion: "16214"
  selfLink: /apis/sources.kyma-project.io/v1alpha1/namespaces/kyma-integration/httpsources/varkes
  uid: def21dc2-1d1b-4ba0-bfb7-83e20d83c7ef
spec:
  source: varkes
status:
  SinkURI: http://varkes-kn-channel.kyma-integration.svc.cluster.local
  conditions:
  - lastTransitionTime: "2020-02-28T01:31:06Z"
    status: "True"
    type: Deployed
  - lastTransitionTime: "2020-02-28T01:30:43Z"
    severity: Info
    status: "True"
    type: PolicyCreated
  - lastTransitionTime: "2020-02-28T01:31:06Z"
    status: "True"
    type: Ready
  - lastTransitionTime: "2020-02-28T01:30:42Z"
    status: "True"
    type: SinkProvided
```
##  Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

| Parameter   |      Required      |  Description |
|----------|:-------------:|------|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **metadata.namespace** | Yes | Specifies the Namespace in which the CR is created. |
| **spec.source** | Yes | Specifies a human-readable name of the Application service. |
| **status.sinkURI** | Yes | Specifies the endpoint the Application sends the events to. This HTTP source can receive events from this Application only. |