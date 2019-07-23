---
title: Provide Service Classes documentation
type: Details
---

Using the Helm Broker, you can provision an addon which defines a Service Broker and provides its own Service Classes. If you want to provide documentation for those Service Classes, create the `docs.yaml` file inside the addon's chart to comply with the naming convention.
The structure of this file looks as follows:

```yaml
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
If your addon defines a ServiceBroker, use the DocsTopic type. If your addon defines a ClusterServiceBroker, use the ClusterDocsTopic type.
For more details, see the [ClusterDocsTopic](/components/headless-cms/#custom-resource-clusterdocstopic)
and [DocsTopic](/components/headless-cms/#custom-resource-docstopic) custom resources documentation.
For more information about currently supported types of the assets, read [this](/components/headless-cms/#overview-overview-headless-cms-in-kyma) document.

One ClusterDocsTopic or DocsTopic object corresponds to a single Service Class with the same ID as the name of the specified object. Store documentation for each Service Class in the `docs/{class_name}` directory which corresponds to the value of the **filter** parameter in the ClusterDocsTopic or DocsTopic definition.

During the provisioning process, the Helm Broker pushes the **addonRepositoryURL** variable into the chart. The **addonsRepositoryURL** points to your addon compressed to a `.tgz` file.

### Documentation structure

Deliver documentation for your addons in Markdown files with specified metadata. To learn more about the metadata and content of the Markdown files, read [this](/components/headless-cms/#details-markdown-documents) document.
