---
title: Overview
type: Overview
---

Kyma is the easiest and fastest way to integrate and extend products in a cloud-native way. Kyma is designed as a centerpiece that brings together different external products and increases their agility and customizability.

Kyma allows you to extend and customize the functionality of your products in a quick and modern way, using serverless computing and microservice architecture. The extensions and customizations you create are decoupled from the core applications, which means that deployments are quick, scaling is independent from the core applications, and the changes you make can be easily reverted without causing downtime of the production system.

Living outside of the core product, Kyma allows you to be completely language-agnostic and customize your solution using the technology stack you want to use, not the one the core product dictates. Additionally, Kyma follows the "batteries included" principle and comes with all of the "plumbing code" ready to use, allowing you to focus entirely on writing the domain code and business logic.

Out of the box, Kyma comes with:
  - Security (Service Identity, TLS, Role Based Access Control)
  - Resilience
  - Telemetry and reporting
  - Traffic routing
  - Fault injection

When it comes to technology stacks, Kyma is all about the latest, most modern, and most powerful technologies available. The entire solution is containerized and runs on a [Kubernetes](https://kubernetes.io/) cluster hosted in the [Microsoft Azure](https://azure.microsoft.com/) cloud environment. Customers can access the cluster easily using a single sign on solution based on the [Dex](https://github.com/coreos/dex) identity provider integrated with any [OpenID Connect](https://openid.net/connect/)-compliant identity provider or a SAML2-based enterprise authentication server.

The communication between services is handled by the [Istio](https://istio.io/) service mesh component, which enables security, monitoring, and tracing without the need to change the application code.
Build your applications using services provisioned by one of the many Service Brokers compatible with the [Open Service Broker API](https://www.openservicebrokerapi.org/), and monitor the speed and efficiency of your solutions using [Prometheus](https://prometheus.io/), which gives you the most accurate and up-to-date tracing and telemetry data.

<p class="internal">
Using [Minikube](https://github.com/kubernetes/minikube), you can run Kyma locally, develop, and test your solutions on a small scale before you push them to a cluster. Follow the Getting Started guides to [install Kyma locally](031-gs-local-installation.md) and [deploy a sample service](032-gs-sample-service-deployment-to-local.md).</p>
