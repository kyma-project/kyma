---
title: Service Classes documentation provided by bundles
type: Details
---

Using the Helm Broker, you can provision a bundle which provides its own Service Classes. If you want to provide documentation for those Service Classes, create the `docs.yaml` file inside the bundle's chart. The structure of this file looks as follows:

```
apiVersion: cms.kyma-project.io/v1alpha1
kind: DocsTopic
metadata:
  labels:
    cms.kyma-project.io/view-context: service-catalog
  name: {ServiceClass ID}
spec:
  displayName: {displayName}
  description: {description}
  sources:
    - type: {type}
      name: {name}
      mode: {mode}
      url: {{ .Values.addonsRepositoryURL }}
      filter: docs/{class_name}/ 
___
apiVersion: cms.kyma-project.io/v1alpha1
kind: DocsTopic
metadata ... 
```
For detailed descriptions of all parameters, see the [ClusterDocsTopic custom resource](/components/headless-cms/#custom-resource-clusterdocstopic). 
For more information about currently supported types of the assets, read [this](/components/headless-cms/#overview-overview-headless-cms-in-kyma) document.


One ClusterDocsTopic object corresponds to a single Service Class with the same ID as the name of the specified ClusterDocsTopic.
Place all ClusterDocsTopic definitions for your bundle in the `docs.yaml` file.
Store documentation for each Service Class in the `docs/{class_name}` directory which corresponds to the value of the **filter** parameter in the ClusterDocsTopic definition.


During the provisioning process, the Helm Broker pushes the **addonRepositoryURL** variable into the chart. The **addonsRepositoryURL** points to your bundle compressed to a `.tgz` file.


### Documentation structure

Deliver documentation for your bundle in the Markdown files with the specified metadata. The metadata must contain the **title** and **type** fields:

```
title: Services and Plans
type: Details
```

The **title** field defines the title of the document displayed in the Console.