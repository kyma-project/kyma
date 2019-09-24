---
title: Overview
---

>**NOTE:** Compass is a new, experimental component in Kyma. To learn how to enable it, read the [installation](#installation-enable-compass-in-kyma) document.

Compass is a multi-tenant system that allows you to connect external applications to Kyma, and manage those applications across multiple [Kyma Runtimes](#architecture-components-kyma-runtime). Using Compass, you can control and monitor your application landscape in one central place. As an integral part of Kyma, Compass uses a set of Kyma features, such as Istio, Prometheus, Monitoring, or Tracing. It also contains Compass UI Cockpit that exposes Compass APIs to users.
These are the functionalities that Compass provides:
- Connect external applications to Kyma
- Connect and manage multiple Kyma Runtimes in one central place
- Integrate applications with Runtimes
- Store applications and Runtimes configurations
- Communicate the configuration changes to applications and Runtimes
- Establish a trusted connection between applications and Runtimes

Compass itself does not store applications' or Runtimes' business logic and does not participate in any business flow. In case of any connection error, the applications and Runtimes workflow does not break and there is only a disruption in transmitting a new configuration.
