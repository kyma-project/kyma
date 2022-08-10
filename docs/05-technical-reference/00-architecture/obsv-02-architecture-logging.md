---
title: Logging Architecture
---

## Architecture diagram

![Logging architecture in Kyma](./assets/obsv-logging-architecture.drawio.svg)

## Process flow

1. Container logs are stored under the `var/log` directory and its subdirectories.
2. The agent detects any new log files in the folder and tails them.
3. The agent queries the [Kubernetes API Server](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/) for additional Pod metadata, such as Pod annotations and labels.
4. The agent enriches log data with labels and sends them to the Loki server.
5. The Loki server processes the log data and stores it in the log store. The data is indexed based on the passed labels
6. The user queries the logs using Grafana dashboards to analyze and visualize logs fetched and processed by Loki. Learn more about [accessing Grafana](../../04-operation-guides/security/sec-06-access-expose-kiali-grafana.md).

## Telemetry Component

![](./assets/obsv-configurable-logging-architecture.drawio.svg)

Kyma's classic in-cluster logging features (1-6) are unchanged, but with the integration of Kyma's [telemetry component](./../../01-overview/main-areas/observability/obsv-04-telemetry-in-kyma.md), you can use additional functionality:

7. The telemetry component provides your custom output configuration for Fluent Bit.
8. As specified in your configuration, Fluent Bit sends the log data to observability systems outside the Kyma cluster.
9. The user accesses the external observability system to analyze and vizualize the logs.

Learn how to [enable the telemetry component](../../04-operation-guides/operations/obsv-00-enable-telemetry_component.md).