---
title: Observability
---

We successfully got our Function [triggered](04-trigger-workload-with-event.md).
Now we would like to check its logs and metrics. 
To do that, we'll use the Grafana dashboard that comes with Kyma.

> **NOTE:** See how to access logs from the Function's Pod [via Kyma Dashboard](../04-operation-guides/operations/obsv-01-access-logs.md#kubernetes-logs-in-kyma-dashboard) and [using kubectl](../04-operation-guides/operations/obsv-01-access-logs.md#kubernetes-logs-using-kubectl). 

## Access Grafana

> **NOTE:** See how to [expose Grafana securely](../04-operation-guides/security/sec-06-access-expose-kiali-grafana.md) for easier access in the future.

1. To access Grafana, forward a local port to the Service's port:
    ```bash
    kubectl -n kyma-system port-forward svc/monitoring-grafana 3000:80
    ```
2. In your browser, go to [`http://localhost:3000`](http://localhost:3000) to open Grafana dashboard.

## View the logs

<div tabs name="View the logs" group="view-logs">
  <details open>
  <summary label="Grafana dasboard">
  Grafana dashboard
  </summary>

1. Use the left menu to navigate to **Explore** and choose **Loki** from the dropdown list.
2. Click on **Log browser** and select the following values:
   - **1. Select labels to search in**: `container`, `function`
   - **2. Find values for the selected labels**: for **function** choose `lastorder`, for **container** choose `function`
  
3. Click **Show logs**.

    > **NOTE:** Alternatively, type or paste the `{function="lastorder", container="function"}` query and press `Shift`+`Enter` or click on **Run query**. 

You can now browse the logs.

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. Go to the `default` Namespace.
   
2. Go to **Workloads** > **Functions**.
   
3. Select your `lastorder` Function and click on **View Logs**.

You can now browse the logs.
  </details>
</div>

## View the metrics

1. In the Grafana dashboard, use the search tool from the left menu.
2. Search for the `Kyma / Function` board and select it. 
3. From the **Function** dropdown, select `lastorder`.

You can now view the metrics, such as success rate and resource consumption.

That's it! 

Go ahead and dive a little deeper into the Kyma documentation for [tutorials](../03-tutorials), [operation guides](../04-operation-guides), and [technical references](../05-technical-reference), as well as information on the [main areas in Kyma](../01-overview/main-areas). Happy Kyma-ing!