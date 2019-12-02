---
title: Architecture
---

This document provides an overview of the logging architecture in Kyma. 

![Logging architecture in Kyma](./assets/logging-architecture.svg)



1. The container logs are stored in the file system under the `var/log` directory.
2. The Agent discovers Pods by querying the [Kubernetes API Server](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/) which validates and configures data for objects such as Pods or Services. Thanks to this configuration the Agent knows which containers belong to a particular Pod and can tail them. 
3. The Agent enriches the log data with Pod labels and sends them to the Loki server. The log data is organized in log chunks. A log chunk consists of all logs for metadata, such as labels, collected over a certain time period. Log chunks support append, seek, and stream operations on requests.
4. The Loki server processes and stores log data in the log store. The Pod labels are stored in index store and used for filtering
5. The user can query the logs in the following tools:

* Grafana dashboard used to filter and display the logs.
* API client to query the log data using the [HTTP API](https://github.com/grafana/loki/blob/master/docs/api.md).
* Log UI accessed from the Kyma Console.






