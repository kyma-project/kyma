---
title: Access Kyma application logs
---

To get insights into your applications (microservices, Functions...), you view the respective logs. You can check out real-time logs immediately using the Kubernetes functionalities. If you want to see historical logs and use the search function, use the logs as aggregated by Loki.
Kyma's logging stack provides additional features: In Grafana, you can see a visual representation of your logs and filter functions for up to five days. If you want to see logs from up to five days back and to use search functionality, call the Loki API directly.

## Kubernetes logs via kubectl

```bash
kubectl logs {POD_NAME} --namespace {NAMESPACE_NAME} --container {CONTAINER_NAME}
```

## Kubernetes logs via Kyma Dashboard

Select the namespace, access the Pod, select the container and use **View Logs**.

## Loki logs via Grafana UI

In Kyma Dashboard, go to Observability and open Grafana.

In Grafana's **Explore** section, select `Loki` as data source and enter a query following the [guidelines](https://grafana.com/docs/loki/latest/logql/), for example:

```bash
{namespace="kyma-system"} |= "info"
```

## Loki logs via Loki API

To access the logs through the Loki API directly, follow these steps:

1. Configure port forwarding with the following command:

   ```bash
   kubectl port-forward -n kyma-system service/logging-loki 3100:3100
   ```

2. To get first 1000 lines of info logs for components in the `kyma-system` Namespace, run the following command:

   ```bash
   curl -X GET -G 'http://localhost:3100/api/prom/query' --data-urlencode 'query={namespace="kyma-system"}' --data-urlencode 'limit=1000' --data-urlencode 'regexp=info'
   ```
