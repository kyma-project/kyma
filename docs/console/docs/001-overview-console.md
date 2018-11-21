---
title: Overview
type: Overview
---

The Console is a web-based administrative UI for Kyma. It allows you to administer the Kyma functionality and manage the basic Kubernetes resources.

The Console uses the Luigi framework to extend the UI functionality with custom micro front-ends and bring more functionality to the existing UI. You can define the micro front-ends using dedicated CustomResourceDefinitions (CRDs).

Use the following CRs to extend the Console UI:

- The MicroFrontend custom resource allows you to plug in micro front-ends for a specific Environment.
- The ClusterMicroFrontend custom resource allows you to plug in micro front-ends for the entire Cluster.
