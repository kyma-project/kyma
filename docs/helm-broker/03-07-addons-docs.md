---
title: Provide documentation for your addon
type: Details
---

Using the Helm Broker, you can provide documentation for your addon and display it in the Console UI. There are two cases in which you may want to use this feature:

- Providing documentation for your addon
- Providing documentation for objects that appear after provisioning your addon

>**NOTE:** Deliver documentation for your addons in Markdown files with specified metadata. To learn more about the metadata and content of the Markdown files, read [this](/components/service-catalog/#console-ui-views-specifications-in-the-console-ui-markdown-documents) document. For more information about the currently supported types of assets, read [this](/components/rafter/#overview-overview-rafter-in-kyma) document.
<!-- Check if the links work once Rafter is already in Kyma. -->

## Provide documentation for your addon

To provide documentation for your addon, create the `docs` folder inside your addon's directory. Your `docs` folder must contain a [`meta.yaml`](#details-create-addons-docs-directory) file with metadata information about how documentation for the addon is uploaded. You can either provide your own documents or point to the external URL with the source documentation:

<div tabs name="provide-documentation-for-your-addon" group="provide-documentation-for-your-addon">
  <details>
  <summary label="provide-your-own-documentation">
  Provide your own documentation
  </summary>

Store your documents and assets in the `docs` folder inside your addon's directory. Each Markdown file appears in the **Documentation** tab in the Console UI. Point the **filter** parameter of your `meta.yaml` file to the `docs` directory that contains the documentation.

  </details>
  <details>
  <summary label="use-external-documentation">
  Use external documentation
  </summary>

In the `meta.yaml` file, provide the **url** parameter with a value that points to the address of the documentation repository.

  </details>
</div>

## Provide documentation for objects

To provide documentation for objects that appear after provisioning your addon, create the `docs.yaml` file inside the addon's chart. This file contains [ClusterAssetGroup](/components/rafter/#custom-resource-cluster-asset-group) or [AssetGroup](/components/rafter/#custom-resource-asset-group) custom resources. Each ClusterAssetGroup or AssetGroup corresponds to a single object with the same ID as the name of the specified object. Your `docs.yaml` file can contain many ClusterAssetGroups or AssetGroups.

<div tabs name="provide-documentation-for-objects" group="provide-documentation-for-your-addon">
  <details>
  <summary label="provide-your-own-documentation">
  Provide your own documentation
  </summary>

Store documentation for each object in the `docs/{object_name}` directory. In the `docs.yaml` file, set the **url** parameter to the `{{ .Values.addonsRepositoryURL }}` variable, which points to your addon compressed to a `.tgz` file. During the provisioning process, the Helm Broker pushes this variable into the chart. The **filter** parameter in the ClusterAssetGroup or AssetGroup definition must point to the `docs/{object_name}` directory that contains the documentation.

  </details>
  <details>
  <summary label="use-external-documentation">
  Use external documentation
  </summary>

In your `docs.yaml` file, specify the **url** parameter of every ClusterAssetGroup or AssetGroup custom resource with the URL that points to the location containing the documentation for the given object.

  </details>
</div>
