---
Migration Guide 2.19-2.20
---

Prometheus and Grafana have been removed with Kyma 2.20. After upgrading to Kyma 2.20, run the script [2.19-2.20-cleanup-monitoring.sh](./assets/2.19-2.20-cleanup-monitoring.sh) to remove Prometheus and Grafana.

> **NOTE** If you want to continue using Prometheus and Grafana, you can [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus). 
