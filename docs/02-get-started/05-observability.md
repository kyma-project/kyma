---
title: Observability
---

We successfully got our Function [triggered](04-trigger-workload-with-event.md).
Now we would like to check its logs and metrics. 
To do that, we'll use the Grafana dashboard that comes with Kyma.

> **NOTE:** See how to access logs from the Function's Pod [via Kyma Dashboard](../04-operation-guides/operations/obsv-01-access-logs.md#kubernetes-logs-in-kyma-dashboard) and [using kubectl](../04-operation-guides/operations/obsv-01-access-logs.md#kubernetes-logs-using-kubectl).

> **NOTE:** Prometheus and Grafana are [deprecated](https://kyma-project.io/blog/2022/12/9/monitoring-deprecation) and are planned to be removed. If you want to install a custom stack, take a look at [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus).

## Access Grafana

> **NOTE:** See how to [expose Grafana securely](../04-operation-guides/security/sec-06-access-expose-grafana.md) for easier access in the future.

1. To access Grafana, forward a local port to the Service's port:
    ```bash
    kubectl -n kyma-system port-forward svc/monitoring-grafana 3000:80
    ```
<!-- markdown-link-check-disable-next-line -->
2. In your browser, go to [`http://localhost:3000`](http://localhost:3000) to open Grafana dashboard.

## View the metrics

1. In the Grafana dashboard, use the search tool from the left menu.
2. Search for the `Kyma / Function` board and select it. 
3. From the **Function** dropdown, select `lastorder`.

You can now view the metrics, such as success rate and resource consumption.

That's it! 

Go ahead and dive a little deeper into the Kyma documentation for [tutorials](../03-tutorials), [operation guides](../04-operation-guides), and [technical references](../05-technical-reference), as well as information on the [main areas in Kyma](../01-overview/). Happy Kyma-ing!