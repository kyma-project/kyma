---
title: Overview
---

To enrich Kyma with monitoring functionality, third-party resources come by default as packaged tools. The `kube-prometheus` package is a Prometheus operator from CoreOS responsible for delivering these tools. Monitoring in Kyma includes three primary elements:

* Prometheus, an open-source system monitoring toolkit.
* Grafana, a user interface that allows you to query and visualize statistics and metrics.  
* AlertManager, a Prometheus component that handles alerts that originate from Prometheus. AlertManager performs needed deduplicating, grouping, and routing based on rules defined by the Prometheus server.

Convenience and efficiency are the main advantages to using the `kube-prometheus` package. `kube-prometheus` delivers a level of monitoring options that would otherwise involve extensive development effort to acquire. Prometheus, Grafana, and AlertManager installed on their own would require the developer to perform customization to achieve the same results as the operator alone. `kube-prometheus` is configured to run on Kubernetes and monitor clusters without additional configuration.
