---
title: Headless CMS in the Console
type: Details
---

Headless CMS in Kyma is a separate component and does not provide templates or a standalone UI to display content from the CMS. Still, the Console UI displays some of the assets in the Docs view (generic and component-related documentation) and the Service Catalog views (API specifications).

Add a new documentation topic to the Console UI by creating a CR, while the [Asset Store](/components/asset-store/#overview-overview) handles the rest. Define the location, grouping, and order of the documents through DocsTopic and ClusterDocsTopic labels that you add to the custom resource definition. You can create your own labels and follow your own naming convention, but **cms.kyma-project.io/{label-name}** applies to the Console UI.

Use these labels to mark and filter assets in the Console UI:

- **cms.kyma-project.io/view-context** specifies the location in the Console UI to render the given asset. This can be either `docs-ui` or `service-catalog`.
- **cms.kyma-project.io/group-name** defines the group, such as `components`, under which you want to render the given asset under the `docs-ui` view in the Console UI. The value cannot include spaces.
- **cms.kyma-project.io/order** specifies the position of the DocsTopic and ClusterDocsTopic in relation to other DocsTopics under the `docs-ui` view in the Console UI. For example, this can be `4`.

## Configuration

To configure how DocsTopic and ClusterDocsTopic are rendered in UI, use the following parameters:

| Parameter | Default Value | Description |
| --------- | ------------- | ----------- |
| **spec.sources.metadata.disableRelativeLinks** | `false` | Disables relative links when documentation is rendered. |
