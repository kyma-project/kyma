---
title: Rafter Controller Manager
type: Metrics
---

Metrics for the Rafter Controller Manager include:

- default metrics instrumented by [kubebuilder](https://book.kubebuilder.io/).
- default Prometheus metrics for [Go applications](https://prometheus.io/docs/guides/go-application/).

To see a complete list of metrics, run this command:

```bash
kubectl -n kyma-system port-forward svc/rafter-controller-manager 8080
```

To check the metrics, open a new terminal window and run:

```bash
curl http://localhost:8080/metrics
```

> **TIP:** To use these commands, you must have a running Kyma cluster and kubectl installed. If you cannot access port `8080`, redirect the metrics to another one. For example, run `kubectl -n kyma-system port-forward svc/rafter-controller-manager 3000:8080` and update the port in the localhost address.

See the [Monitoring](/components/monitoring) documentation to learn more about monitoring and metrics in Kyma.
