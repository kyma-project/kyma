---
title: Technology stack
type: Overview
---

The entire solution is containerized and runs on a [Kubernetes](https://kubernetes.io/) cluster. Customers can access it easily using a single sign on solution based on the [Dex](https://github.com/coreos/dex) identity provider integrated with any [OpenID Connect](https://openid.net/connect/)-compliant identity provider or a SAML2-based enterprise authentication server.

The communication between services is handled by the [Istio](https://istio.io/) Service Mesh component which enables security, traffic management, routing, resilience (retry, circuit breaker, timeouts), monitoring, and tracing without the need to change the application code.
Build your applications using services provisioned by one of the many Service Brokers compatible with the [Open Service Broker API](https://www.openservicebrokerapi.org/), and monitor the speed and efficiency of your solutions using [Prometheus](https://prometheus.io/), which gives you the most accurate and up-to-date monitoring data.
