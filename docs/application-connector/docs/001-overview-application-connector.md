---
title: Overview
---

The Application Connector (AC) is a proprietary Kyma implementation that allows you to connect with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process doesn't change, which allows to avoid any configuration or network-related problems.

The external solution you connect to Kyma using the AC is represented as an Application (App). There is always a one-to-one relationship between a connected solution and an App, which helps to ensure the highest level of security and separation. This means that you must create five separate Apps in your cluster to connect five different external solutions and use their APIs and Event catalogs in Kyma.

The AC gives you this functionality:

- Manages the lifecycle of Apps.
- Establishes a secure connection and generates the client certificate used by the connected external solution.
- Registers the APIs and the Event catalogs of the connected external solution.
- Delivers the Events from the connected external solution to the Kyma Event Bus.
- Proxies calls sent from Kyma to external APIs registered by the connected external solution.
- Allows to map an App to a Kyma Namespace and use its registered APIs and Event catalogs in the context of that Namespace.
- Integrates the registered APIs and Event catalogs with the Kyma Service Catalog.

All of the AC components scale independently, which allows to adjust it to fit the needs of the implementation built using Kyma.

>**NOTE:** To learn more about the concept of Namespaces in Kyma, read the **Namespaces** document in the **Kyma** documentation topic.
