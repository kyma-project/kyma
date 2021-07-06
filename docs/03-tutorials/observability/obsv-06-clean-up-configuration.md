---
title: Clean up the configuration
---

## Purpose

When you're finished working with the monitoring example, remove the example deployment and all its resources from the cluster.

> **NOTE:** Do not clean up the resources if you want to continue with the following tutorial as these resources are used there as well.

## Steps

1. Remove the deployed ServiceMonitor CRD from the `kyma-system` Namespace.

    ```bash
    kubectl delete servicemonitor -l example=monitoring-custom-metrics -n kyma-system
    ```

2. Remove the example deployment from the `testing-monitoring` Namespace.

    ```bash
    kubectl delete all -l example=monitoring-custom-metrics -n testing-monitoring
    ```

3. Remove the `testing-monitoring` Namespace.

    ```bash
    kubectl delete namespace testing-monitoring
    ```