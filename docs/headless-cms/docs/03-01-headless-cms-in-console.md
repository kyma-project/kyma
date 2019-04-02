---
title: Headless CMS in Console
type: Details
---

Headless CMS in Kyma is a separate component and does not provide templates or a standalone UI to display content from CMS. Still, some of its assets are shown in the Console UI in the Docs view (generic and component-related documentation) and the Service Catalog view (API specifications).

You add a new documentation topic to the Console UI by creating a CR, while the [Asset Store](#asset-store-overview) handles the rest. You define the location, grouping, and order of the documents through these DocsTopic and ClusterDocsTopic labels:

- **cms.kyma-project.io/view-context** that specifies the location in the Console UI where you want to render the given asset. This can be either `docs-ui` or `service-catalog`.
- **cms.kyma-project.io/group-name** that defines the group, such as `components`, under which you want to render the given asset in the Console UI. The value cannot include spaces.
- **cms.kyma-project.io/order** that specifies the position of the DocsTopic and ClusterDocsTopic in the Console UI view in relation to other DocsTopics. For example, this can be `4`.

>**NOTE:** Labels in the Docs and Catalog views in the Console UI follow the **cms.kyma-project.io/{label-name}** naming convention to make filtering easier. Use your own naming convention for labels applied in different contexts.
