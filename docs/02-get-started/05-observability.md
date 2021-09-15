---
title: Observability
---

We successfully got our Function [triggered](04-trigger-workload-with-event.md).
Now we would like to check its logs. 
To do that, we'll use the Grafana dashboard that comes with Kyma. 

> **NOTE:** See how to access logs from the Function's Pod [via Kyma Dashboard](../04-operation-guides/operations/obsv-01-access-logs.md#kubernetes-logs-in-kyma-dashboard) and [using kubectl](../04-operation-guides/operations/obsv-01-access-logs.md#kubernetes-logs-using-kubectl). 

1. To access Grafana, forward a local port to a port on the service's Pod:
    ```bash
    kubectl -n kyma-system port-forward svc/monitoring-grafana 3000:80
    ```
2. In your browser, go to [`http://localhost:3000`](http://localhost:3000) to open Grafana dashboard.
3. Using the left menu, navigate to **Explore** and choose **Loki** from the dropdown list.
4. Click on **Log browser** and select the following values:
   - **1. Select labels to search in**: `container`, `function`
   - **2. Find values for the selected labels**: for **function** choose `lastorder`, for **container** choose `function` 
   
   Click **Show logs**.

    > **NOTE:** Alternatively, type or paste the `{function="lastorder", container="function"}` query and press `Shift`+`Enter` or click on **Run query**. 

You may now browse the logs. That's it! 

Go ahead and dive a little deeper into the Kyma documentation for more [guides](../04-operation-guides), [tutorials](../03-tutorials), and [technical references](../05-technical-reference), as well as information on the [main areas in Kyma](../01-overview/main-areas). Happy Kyma-ing!