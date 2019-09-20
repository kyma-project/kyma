---
title: Provide documentation for your addon
type: Details
---

Using the Helm Broker, you can provide documentation for your addon and display it in the Console UI. There are two cases in which you may want to use this feature:
- Providing documentation for your addon
- Providing documentation for objects that appear after provisioning your addon

>**NOTE:** Deliver documentation for your addons in Markdown files with specified metadata. To learn more about the metadata and content of the Markdown files, read [this](/components/headless-cms/#details-markdown-documents) document. For more information about the currently supported types of assets, read [this](/components/headless-cms/#overview-overview-headless-cms-in-kyma) document.

## Provide documentation for your addon

To provide documentation for your addon, create the `docs` folder inside your addon's directory. Your `docs` folder must contain a [`meta.yaml`](#details-create-addons-docs-directory) file with metadata information about how documentation for the addon is uploaded. You can either provide your own documents or point to the external URL with the source documentation:

<div tabs>
  <details>
  <summary>
  Provide your own documentation
  </summary>

Store your documents and assets in the `docs` folder inside your addon's directory. Each Markdown file represents a separate tab in the Console UI. The **type** metadata of your Markdown documents determines the order of the documents. Point the **filter** parameter of your `meta.yaml` file to the `docs` directory that contains the documentation.

  </details>
  <details>
  <summary>
  Use external documentation
  </summary>

In the `meta.yaml` file, provide the **url** parameter with a value that points to the address where the documentation is stored.

  </details>
</div>

## Provide documentation for objects

To provide documentation for objects that appear after provisioning your addon, create the `docs.yaml` file inside the addon's chart. This file contains [ClusterDocsTopic](/components/headless-cms/#custom-resource-cluster-docs-topic) or [DocsTopic](/components/headless-cms/#custom-resource-docstopic) custom resources. Each ClusterDocsTopic or DocsTopic corresponds to a single object with the same ID as the name of the specified object. Your `docs.yaml` file can contain many ClusterDocsTopics or DocsTopics.

<div tabs>
  <details>
  <summary>
  Provide your own documentation
  </summary>

Store documentation for each object in the `docs/{object_name}` directory. In the `docs.yaml` file, set the **url** parameter to the `{{ .Values.addonsRepositoryURL }}` variable, which points to your addon compressed to a `.tgz` file. During the provisioning process, the Helm Broker pushes this variable into the chart. The **filter** parameter in the ClusterDocsTopic or DocsTopic definition must point to the `docs/{object_name}` directory that contains the documentation.

  </details>
  <details>
  <summary>
  Use external documentation
  </summary>

In your `docs.yaml` file, specify the **url** parameter of every ClusterDocsTopic or DocsTopic custom resource with the URL that points to the location containing the documentation for a given object.

  </details>
</div>
