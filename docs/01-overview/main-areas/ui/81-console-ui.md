---
title: Console UI
---

The Console is a web-based administrative UI for Kyma. It uses the [Luigi](https://github.com/SAP/luigi) framework to allow you to seamlessly extend the UI content with custom micro frontends. The Console has a dedicated Console Backend Service which provides a tailor-made API for each view of the Console UI.

Use the following CRs to modify the Console UI:
- The MicroFrontend custom resource allows you to plug in micro frontends for a specific Namespace.
- The ClusterMicroFrontend custom resource allows you to plug in micro frontends for the entire Cluster.
- The BackendModule custom resource allows you to enable Console Backend Service modules.

Learn how to [extend the Console UI](../../../03-tutorials/tut-extend-ui-luigi.md) with custom micro frontends.
