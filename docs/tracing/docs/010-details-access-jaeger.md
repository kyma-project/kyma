---
title: Access Jaeger
type: Details
---

To access the Jaeger UI, follow these steps:

1. Run the following command to configure port-forwarding:

```
kubectl port-forward -n kyma-system $(kubectl get pod -n kyma-system -l app=jaeger -o jsonpath='{.items[0].metadata.name}') 16686:16686
```

2. Access the Jaeger UI at `http://localhost:16686`.
