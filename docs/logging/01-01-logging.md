---
title: Overview
---

Logging in Kyma uses [Loki](https://github.com/grafana/loki), which is a lightweight Prometheus-like log management system. The main elements of the logging stack include:

* The Agent acting as a log router for Docker containers. It runs inside Docker, checks each container, and routes the logs to the log management system. Currently, Kyma supports the [Fluent Bit](https://fluentbit.io/) log collector.
* Loki main server, which stores logs and processes queries.
* [Grafana](https://grafana.com/), which queries and displays logs.

>**CAUTION:** Loki is designed for application logging. Do not log any sensitive information, such as passwords or credit card numbers.

You can install Logging as part of Kyma predefined [profiles](/root/kyma/#installation-overview-profiles). For production purposes, use the **production profile** which has higher memory limits set for Loki and Fluent Bit logging solutions. It ensures stable log processing for 40 active Pods without causing any memory issues. If you want to work with a larger number of active Pods or experience a prolonged query time, [configure the Logging chart](#configuration-logging-chart) to increase the memory and CPU values.

**>NOTE:** The production profile does not allow for horizontal scaling for which you need additional, dedicated storage systems.
