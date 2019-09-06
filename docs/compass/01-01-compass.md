---
title: Overview
---

>**NOTE:** Compass is a new, experimental component in Kyma. To enable it, follow the instructions for the [custom component installation](/root/kyma#configuration-custom-component-installation) and enable the `compass` and `compass-runtime-agent` modules.

Compass is a multi-tenant system which consists of components that provide a way to manage your applications across multiple Kyma Runtimes. Using Compass, you can control and monitor your application landscape in one central place.

Compass allows for registering different types of applications and Runtimes.
These are the types of possible integration levels between an application and Compass:
- Manual integration - administrator manually provides API/Events metadata to Compass. This type of integration is used mainly for simple use-case scenarios and doesn't support all features.
- Built-in integration - integration with Compass is built-in inside the application.
- Proxy - a highly application-specific proxy component provides the integration.
- Central integration service -  a central service provides integration for the whole class of applications. It manages multiple instances of these applications. You can integrate multiple central services to support different types of applications.

Runtime is any system that can configure itself according to the configuration provided by Compass. You can register any Runtime, providing that it fulfills a contract with Compass and implements its flow. First, your Runtime must get a trusted connection to Compass. It must also allow for fetching application definitions and using these applications in a given tenant. By default, Compass is integrated with Kyma (Kubernetes), but its usage can also be extended to other platforms, such as CloudFoundry or Serverless.

Compass, as a part of Kyma, uses a set of Kyma features, such as Istio, Prometheus, Monitoring, or Tracing. It also contains Compass UI Cockpit that exposes Compass APIs to users.

## Run Kyma with Compass

You can run Kyma with Compass in two modes:

- Default Kyma installation which provides all Kyma components together with Compass and Agent, which allows you to register external applications to Kyma.

![Kyma mode2](./assets/kyma-mode2.svg)

- Kyma as a central solution which allows you to connect and manage your multiple Kyma Runtimes. It consists of Compass and only the selected Kyma components.

![Kyma mode1](./assets/kyma-mode1.svg)

For more details, read the [architecture](#architecture-architecture) documents.
