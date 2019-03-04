---
title: Access logs
type: Details
---

To access the logs, follow these steps:

1. Run the following command to configure port forwarding:
```bash
kubectl port-forward -n kyma-system svc/logging-loki 3100:3100
```

2. To get first 1000 lines of logs for components showing errors in the 'kyma-system' Namespace run the following command:
```bash
curl -X GET 'http://localhost:3100/api/prom/query' -d 'query={namespace="kyma-system"}' -d 'regexp=Error' -d 'limit=1000'
```

For further information, see the [Loki API documentation](https://github.com/grafana/loki/blob/master/docs/api.md).
