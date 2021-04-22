---
title: Kiali
---

Kyma uses [Kiali](https://www.kiali.io) to enable validation, observe the Istio Service Mesh, and provide details on microservices included in the Service Mesh and connections between them.
Kiali offers a set of dashboards and graphs that allow you to have the full Service Mesh at a glance and quickly identify problems and configuration issues.
For more details about particular features, see the [official Kiali documentation](https://kiali.io/documentation/latest/features/).

>**NOTE:** Kiali is disabled by default in Kyma Lite (local installation). Read about [custom component installation](/root/kyma/#configuration-custom-component-installation) for instructions on how to enable it.

You can easily access Kiali from the [Kyma Console](/components/console/#overview-overview). To do so, click the **Service Mesh** tab in the menu on the left.
Once you are authenticated, the main Kiali dashboard will show a summary of the Service Mesh status and the left side menu will offer you features such as graphs or configuration validation:
![Kiali menu item](assets/overview.png)

Use the graphs to review the topology of the Service Mesh:
![Kiali menu item](assets/graph.png)
