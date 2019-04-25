---
title: Service Classes documentation provided by bundles
type: Details
---


The Helm Broker pushes the **addonRepositoryURL** variable into the installed chart during the provisioning process. This feature enables a bundle which provides its own Service Classes, such as the Azure Broker bundle, to also install documentation for them.

If the bundle provides Service Classes and you want to have documentation for them, create a `docs.yaml` file inside the bundle's chart with the following structure:
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
      name: {name}
      mode: {mode}
      url: {.Values.addonsRepositoryURL}
      filter: docs/{class_name}/ 
___
apiVersion: cms.kyma-project.io/v1alpha1
kind: ClusterDocsTopic
metadata ... 
```
where:
      - `name: {name}` must be unique within sources
      - `filter: docs/{class_name}/` is optional. Use it only for the `package` mode.
One ClusterDocsTopic object corresponds to a single Service Class with the same ID as the name of the specified ClusterDocsTopic.

Store the ClusterDocsTopics definitions for a given Service Broker in one file.

If you want to provide documentation for every class provided by the broker you must also provide the ClusterDocsTopics for all of them.

Store the documentation for the Service Classes provided by a bundle in a `docs/{class_name}` folder.
