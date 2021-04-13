---
title: Alertmanager
type: Details
---

Alertmanager receives and manages alerts coming from Prometheus. It can then forward the notifications about fired alerts to specific channels, such as Slack or VictorOps.

## Alertmanager configuration

Use the following files to configure and manage Alertmanager:

* `alertmanager.yaml` which deploys the Alertmanager Pod.
* `values.yaml` which you can use to define core Alertmanager configuration and alerting channels. For details on configuration elements, see the [Prometheus documentation](https://prometheus.io/docs/alerting/configuration/).

## Alerting rules

Kyma comes with a set of [alerting rules](https://github.com/kyma-project/kyma/tree/main/resources/monitoring/templates/prometheus/rules-1.14) provided out of the box.
These rules provide alerting configuration for logging, web apps, rest services, and custom Kyma rules.
You can also define your own alerting rule. To learn how, see the [tutorial](/components/monitoring/#tutorials-define-alerting-rules).
