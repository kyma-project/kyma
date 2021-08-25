---
title: Access Kyma application logs
---

Get insights into your applications (microservices, Functions...) by viewing the respective logs.

To check out real-time logs immediately, use the Kubernetes functionalities - either with the CLI, or in Kyma Dashboard.

If you want to see historical logs and use additional features, view the logs as aggregated by Loki. In Grafana, you can see a visual representation of your logs and filter functions for up to five days. If you want to see logs from up to five days back and to use search functionality, call the Loki API directly.

## Kubernetes logs usingvia kubectl

Run the following command:

```bash
kubectl logs {POD_NAME} --namespace {NAMESPACE_NAME} --container {CONTAINER_NAME}
```

## Kubernetes logs in Kyma Dashboard

1. Open Kyma Dashboard and select the Namespace.
2. Access the Pod and select the container.
3. Click **View Logs**.

## Loki logs in Grafana UI

To see a visual representation of the Loki logs

1. In Kyma Dashboard, go to **Observability** and open **Grafana**.
2. In Grafana's **Explore** section, select `Loki` as data source.
3. Enter a query following the [query language guidelines](https://grafana.com/docs/loki/latest/logql/), for example:

   ```bash
   {namespace="kyma-system"} |= "info"
   ```

## Loki logs using Loki API

To access the logs through the Loki API directly, follow these steps:

1. Configure port forwarding with the following command:

   ```bash
   kubectl port-forward -n kyma-system service/logging-loki 3100:3100
   ```

2. To get first 1000 lines of info logs for components in the `kyma-system` Namespace, run the following command:

   ```bash
   curl -X GET -G 'http://localhost:3100/loki/api/v1/query' --data-urlencode 'query={namespace="kyma-system"}' --data-urlencode 'limit=1000' --data-urlencode 'regexp=info'
   ```
