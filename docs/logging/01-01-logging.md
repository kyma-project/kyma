---
title: Overview
---

Logging in Kyma uses [Loki](https://github.com/grafana/loki) which is a Prometheus-like log management system. This lightweight solution, integrated with Grafana, is easy to understand and operate. The main elements of the logging stack include:
* [Promtail](https://github.com/grafana/loki/tree/master/docs/clients/promtail) which is a log router for Docker containers. Promtail runs inside Docker, checks each container and routes the logs to the log management system.
* Loki main server used to store and logs and process queries.
* [Grafana](https://grafana.com/)logging and metrics platform used for quering and displaying logs. 


>**NOTE:** At the moment, Kyma provides an **alpha** version of the Logging component. The default Loki Pod log tailing configuration does not work with Kubernetes version 1.14 (for GKE version 1.12.6-gke.X) and above. For setup and preparation of deployment see the [cluster installation](/root/kyma/#installation-install-kyma-on-a-cluster) guide.

>**CAUTION:** Loki is designed for application logging. Do not log any sensitive information, such as passwords or credit card numbers.
