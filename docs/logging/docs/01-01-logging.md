---
title: Overview
---

Logging in Kyma uses [Loki](https://github.com/grafana/loki) which is a Prometheus-like log management system. This lightweight solution, integrated with Grafana, is easy to understand and operate. Loki provides Promtail which is a log router for Docker containers. Promtail runs inside Docker, checks each container and routes the logs to the log management system.

> **NOTE:** At the moment, Kyma provides an alpha version of the Logging component.

> **NOTE:** Loki default pod log tailing configuration will not work with kubernetes version 1.14 (for GKE version 1.12.6-gke.X) and above. For setup and preparation of deployment please consult Install Kyma on a cluster guide.

> **NOTE:** Loki designed for application logging, be careful to not log any sensitive information like passwords, credit card numbers and etc.
