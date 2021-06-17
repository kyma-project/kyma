---
title: Access Kyma logs
---

There are two kinds of logs that give you insights into your cluster - Kubernetes logs and Loki logs.

Check out real-time logs immediately using the Kubernetes functionalities.

Kyma's logging stack provides additional features:

In Grafana, you can see a visual representation of your logs and filter functions for up to five days

If you want to see logs from up to five days back and to use search functionality, call the Loki API directly. 


## Kubernetes logs via kubectl


## Kubernetes logs via Busola UI


## Loki logs via Grafana UI


## Loki logs via Loki API

To access the logs through the Loki API directly, follow these steps:

1. Run the following command to get the Pod name:

   ```bash
   kubectl get pods -l app=loki -n kyma-system
   ```

2. Configure port forwarding with the following command, replacing **{pod_name}** with the output of the previous command:

   ```bash
   kubectl port-forward -n kyma-system <pod_name> 3100:3100
   ```

3. To get first 1000 lines of error logs for components in the `kyma-system` Namespace, run the following command:

   ```bash
   curl -X GET -G 'http://localhost:3100/api/prom/query' --data-urlencode 'query={namespace="kyma-system"}' --data-urlencode 'limit=1000' --data-urlencode 'regexp=error'
   ```
