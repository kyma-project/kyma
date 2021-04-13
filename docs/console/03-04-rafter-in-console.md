---
title: Rafter in the Console
type: Details
---

Rafter in Kyma is a separate component and does not provide templates or a standalone UI to display content from the CMS. Still, the Console UI displays some of the assets under the Service Catalog view (API specifications and Service Class-related documentation).

Add a new API specification or Service Class-related documentation topic to the Console UI by creating a CR, while [Rafter](/components/rafter/#overview-overview) handles the rest. Define the location, grouping, and order of the documents through AssetGroup and ClusterAssetGroup labels that you add to the custom resource definition. You can create your own labels and follow your own naming convention, but **rafter.kyma-project.io/{label-name}** applies to the Console UI.

Use the **rafter.kyma-project.io/view-context** label to mark and filter Service Class-related assets in the Console UI. Set its value to `service-catalog` to specify that these assets should be displayed under the Service Catalog view.

## Configuration

To define how AssetGroup and ClusterAssetGroup are rendered in the UI, use the following parameters:

| Parameter | Default Value | Description |
| --------- | ------------- | ----------- |
| **spec.sources.metadata.disableRelativeLinks** | `false` | Disables relative links when documentation is rendered. It only applies to the `markdown` type of assets included in the (Cluster)AssetGroup CR. |

## Supported specifications

The Console UI supports only certain specification types, formats, and versions passed in the AssetGroups and ClusterAssetGroups:

| Type | Description | Format | Version |
| --------- | ------------- | ----------- | ----------- |
| [OpenAPI](https://www.openapis.org/) |   API-related information  | `json` and `yaml`| 3.0 and lower |
| [OData](https://www.odata.org/documentation) |   API-related information  | `xml` | 4.0 and lower |
| [AsyncAPI](https://www.asyncapi.com/) |   Messaging data (for Events)  | `json` and `yaml`| 2.0 and lower |
| Markdown |  Service Class documentation  | `md`|  |

>**NOTE:** OpenAPI, OData, and AsyncAPI specifications rendered in the Console UI follow the [Fiori 3 Fundamentals](https://sap.github.io/fundamental/) styling standards.

The source files are uploaded directly to the given storage without any modifications, except for the following source types:

- `asyncapi` that the AsyncAPI Service validates and, if required, converts to version 2.0 and the `json` format.
- `markdown` from which the Front Matter Service extracts front matter metadata.

![Specification types](./assets/spec-types.svg)

> **TIP:** The default Kyma webhooks that convert and validate `asyncapi` source files and extract metadata from `markdown` files are defined in the [`webhook-config-map.yaml`](https://github.com/kyma-project/kyma/blob/main/resources/rafter/charts/controller-manager/templates/webhooks-config-map.yaml) ConfigMap.
