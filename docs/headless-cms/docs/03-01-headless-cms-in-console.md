---
title: Headless CMS in Console
type: Details
---

Headless CMS in Kyma is a separate component and does not provide templates or a standalone UI to display content from CMS. Still, some of its assets are shown in the Console UI in the Docs view (generic and component-related documentation) and the Service Catalog views (API specifications).

You add a new documentation topic to the Console UI by creating a CR, while the [Asset Store](/components/asset-store/#overview-overview) handles the rest. You define the location, grouping, and order of the documents through DocsTopic and ClusterDocsTopic labels which you add to the custom resource definition. You can create your own labels and follow your own naming preferences, but the **cms.kyma-project.io/{label-name}** naming convention applies in the Console UI.

Use these labels to mark and filter assets in the Console UI:

- **cms.kyma-project.io/view-context** specifies the location in the Console UI to render the given asset. This can be either `docs-ui` or `service-catalog`.
- **cms.kyma-project.io/group-name** defines the group, such as `components`, under which you want to render the given asset under the `docs-ui` view in the Console UI. The value cannot include spaces.
- **cms.kyma-project.io/order** specifies the position of the DocsTopic and ClusterDocsTopic in relation to other DocsTopics under the `docs-ui` view in the Console UI. For example, this can be `4`.
