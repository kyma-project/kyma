---
title: Headless CMS in the Console
type: Details
---

The Headless CMS in Kyma is a separate component and does not provide templates or a standalone UI to display content from the CMS. Still, the Console UI displays some of the assets in the Docs view (generic and component-related documentation) and the Service Catalog views (API specifications and Service Class-related documentation).

Add a new documentation topic to the Console UI by creating a CR, while the [Asset Store](/components/asset-store/#overview-overview) handles the rest. Define the location, grouping, and order of the documents through DocsTopic and ClusterDocsTopic labels that you add to the custom resource definition. You can create your own labels and follow your own naming convention, but **cms.kyma-project.io/{label-name}** applies to the Console UI.

Use these labels to mark and filter assets in the Console UI:

- **cms.kyma-project.io/view-context** specifies the location in the Console UI to render the given asset. This can be either `docs-ui` or `service-catalog`.
- **cms.kyma-project.io/group-name** defines the group, such as `components`, under which you want to render the given asset under the `docs-ui` view in the Console UI. The value cannot include spaces.
- **cms.kyma-project.io/order** specifies the position of the DocsTopic and ClusterDocsTopic in relation to other DocsTopics under the `docs-ui` view in the Console UI. For example, this can be `4`.

## Configuration

To define how DocsTopic and ClusterDocsTopic are rendered in the UI, use the following parameters:

| Parameter | Default Value | Description |
| --------- | ------------- | ----------- |
| **spec.sources.metadata.disableRelativeLinks** | `false` | Disables relative links when documentation is rendered. It only applies to the `markdown` type of assets included in the (Cluster)DocsTopic CR. |

## Supported specifications

The Console UI supports only certain specification types, formats, and versions passed in the DocsTopics and ClusterDocsTopics:

| Type | Description | Format | Version |
| --------- | ------------- | ----------- | ----------- |
| [OpenAPI](https://www.openapis.org/) |   API-related information  | `json` and `yaml`| 3.0 and lower |
| [OData](https://www.odata.org/documentation) |   API-related information  | `xml` | 4.0 and lower |
| [AsyncAPI](https://www.asyncapi.com/) |   Messaging data (for Events)  | `json` and `yaml`| 2.0 and lower |
| Markdown |  Service Class or component documentation  | `md`|  |

>**NOTE:** OpenAPI, OData, and AsyncAPI specifications rendered in the Console UI follow the [Fiori 3 Fundamentals](https://sap.github.io/fundamental/) styling standards.

The source files are uploaded directly to the given storage without any modifications, except for the following source types:

- `asyncapi` that the CMS AsyncAPI Service validates and, if required, converts to version 2.0 and the `json` format.
- `markdown` from which the Asset Metadata Service extracts front matter metadata.

![Specification types](./assets/spec-types.svg)

> **TIP:** The default Kyma webhooks that convert and validate `asyncapi` source files and extract metadata from `markdown` files are defined in the [`webhook-config-map.yaml`](https://github.com/kyma-project/kyma/blob/master/resources/cms/charts/cms-controller-manager/templates/webhook-config-map.yaml) ConfigMap.
