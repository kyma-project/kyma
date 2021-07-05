---
title: Logging Architecture
---

## Architecture diagram

![Logging architecture in Kyma](./assets/obsv-logging-architecture.svg)

## Process flow

1. Container logs are stored under the `var/log` directory and its subdirectories.
2. The agent detects any new log files in the folder and tails them.
3. The agent queries the [Kubernetes API Server](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/) for additional Pod metadata, such as Pod annotations and labels.
4. The agent enriches log data with labels and sends them to the Loki server. To enable faster data processing, log data is organized in log chunks. A log chunk consists of metadata, such as labels, collected over a certain time period.
5. The Loki server processes the log data and stores it in the log store. The data gets indexed on base of the passed labels
6. The user queries the logs using Grafana dashboards to analyze and visualize logs fetched and processed by Loki. 
