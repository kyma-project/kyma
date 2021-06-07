---
title: Overview
---

Kyma comes with third-party applications like Prometheus, Alertmanager, and Grafana, that offer a monitoring functionality for all Kyma resources. These applications are deployed during the Kyma cluster installation, along with a set of pre-defined alerting rules, Grafana dashboards, and Prometheus configuration.

The whole installation package provides the end-to-end Kubernetes cluster monitoring that allows you to:

- View metrics exposed by the Pods.
- Use the metrics to create descriptive dashboards that monitor any Pod anomalies.
- Manage the default alert rules and create new ones.
- Set up channels for notifications informing of any detected alerts.

>**NOTE:** The monitoring component is available by default in the cluster installation, but disabled in the **Kyma Lite** local installation on Minikube. [Enable the component](/root/kyma/#configuration-custom-component-installation-add-a-component) to install it with the [local profile](/components/monitoring/#configuration-monitoring-profiles-local-profile).

You can install Monitoring as part of Kyma predefined [profiles](/root/kyma/#installation-overview-profiles). For production purposes, use the **production profile** which ensures:
- Increased retention time to prevent data loss in case of prolonged troubleshooting.
- Increased memory and CPU values that optimize the performance of your component.
