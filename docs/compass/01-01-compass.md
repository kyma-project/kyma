---
title: Overview
---

>**NOTE:** Compass is a new, experimental component in Kyma. To learn how to enable it, read the [installation](#installation-enable-compass-in-kyma) document.

Compass is a central, multi-tenant system that allows you to connect Applications and manage them across multiple [Kyma Runtimes](#architecture-components-kyma-runtime). Using Compass, you can control and monitor your Application landscape in one central place. As an integral part of Kyma, Compass uses a set of Kyma features, such as Istio, Prometheus, Monitoring, or Tracing. It also contains Compass UI Cockpit that exposes Compass APIs to users.
These are the functionalities that Compass provides:
- Allow to connect and manage Applications and Kyma Runtimes in one central place
- Store Applications and Runtimes configurations
- Group Applications and Runtimes to enable integration
- Communicate the configuration changes to Applications and Runtimes
- Establish a trusted connection between Applications and Runtimes using various authentication methods

Compass by design does not participate in direct communication between Applications and Runtimes. It only allows them to communicate. In case the cluster with Compass is down, the Applications and Runtimes flow still works.
