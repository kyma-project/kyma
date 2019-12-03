---
title: Access logs
type: Details
---

To access the logs, follow these steps:

1. Run the following command to get the Pod name:

   ```bash
   kubectl get pods -l app=loki -n kyma-system
   ```

2. Run the following command to configure port forwarding, replace **{pod_name}** with output of the previous command:

   ```bash
   kubectl port-forward -n kyma-system <pod_name> 3100:3100
   ```

3. To get first 1000 lines of error logs for components in the `kyma-system` Namespace, run the following command:

   ```bash
   curl -X GET -G 'http://localhost:3100/api/prom/query' --data-urlencode 'query={namespace="kyma-system"}' --data-urlencode 'limit=1000' --data-urlencode 'regexp=error'
   ```
