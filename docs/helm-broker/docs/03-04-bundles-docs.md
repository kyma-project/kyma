---
title: Bundles docs
type: Details
---

## Documentation for ServiceClasses provided by bundle

The Helm Broker in the provisioning process pushes the `addonRepositoryURL` variable into installed chart. With this feature a bundle which provides its own ServiceClasses e.g. `Azure Broker` can also install a documentation for them.

If the bundle provides ServiceClasses and you want to have a documentation for it, you should create a `docs.yaml` inside the bundle's chart with the following structure:
```
apiVersion: cms.kyma-project.io/v1alpha1
kind: ClusterDocsTopic
metadata:
  labels:
    cms.kyma-project.io/view-context: service-catalog
  name: {ServiceClass ID}
spec:
  displayName: "Slack Connector Add-on"
  description: "Overall documentation, OpenAPI and AsyncAPI for Slack Connector Add-on"
  sources:
    - type: {type}
      name: {name}  # must be unique within sources
      mode: {mode}
      url: {.Values.addonsRepositoryURL}
      filter: docs/{class_name}/   # provide filter only if you use the package mode
___
apiVersion: cms.kyma-project.io/v1alpha1
kind: ClusterDocsTopic
metadata ... 
```

One ClusterDocsTopic object corresponds to a single ServiceClass with the same ID as the name of specified ClusterDocsTopic.

You should hold the ClusterDocsTopics definitions for a given broker in one file.

If you want to provide documentation for every class provided by the broker you must also provide the ClusterDocsTopics for all of them.

The documentation for the ServiceClasses provided by a bundle should be stored in the `docs/{class_name}` folder.
