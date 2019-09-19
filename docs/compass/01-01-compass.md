---
title: Overview
---

>**NOTE:** Compass is a new, experimental component in Kyma. To learn how to enable it, read the [installation](#installation-enable-compass-in-kyma) document.

Compass is a multi-tenant system which consists of components that provide a way to manage your applications across multiple Kyma Runtimes. Using Compass, you can control and monitor your application landscape in one central place.

Compass allows for registering different types of applications and Runtimes.
These are the types of possible integration levels between an application and Compass:
- Manual integration - the administrator manually provides API or Events metadata to Compass. This type of integration is used mainly for simple use-case scenarios and doesn't support all features.
- Built-in integration - integration with Compass is built in the application.
- Proxy - a highly application-specific proxy component provides the integration.
- Central integration service -  a central service provides integration for the whole group of applications. It manages multiple instances of these applications. You can integrate multiple central services to support different types of applications.

See [this](#architecture-compass-components) diagram for reference.

Runtime is any system to which you can apply configuration provided by Compass. Your Runtime must get a trusted connection to Compass. It must also allow for fetching application definitions and using these applications in a given tenant. By default, Compass is integrated with Kyma (Kubernetes), but its usage can also be extended to other platforms, such as CloudFoundry or Serverless.

As an integral part of Kyma, Compass uses a set of Kyma features, such as Istio, Prometheus, Monitoring, or Tracing. It also contains Compass UI Cockpit that exposes Compass APIs to users.
