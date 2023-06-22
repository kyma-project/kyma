---
title: Observability - Useful links
---

If you're interested in learning more about the Observability area, check out these links:

- Learn how to set up the [Monitoring Flow](../../../03-tutorials/00-observability.md) for your services in Kyma.
- Check out the different ways to [access the logs](../../../04-operation-guides/operations/obsv-01-access-logs.md) provided by Kubernetes and Loki.
- Learn how to adjust Loki's [log limits](../../../04-operation-guides/operations/obsv-02-adjust-loki.md).

- Install a [custom Loki stack](https://github.com/kyma-project/examples/tree/main/loki).
- Install a [custom Jaeger stack](https://github.com/kyma-project/examples/tree/main/jaeger).
- Install a [custom Prometheus stack](https://github.com/kyma-project/examples/tree/main/prometheus).

- To collect and ship workload metrics to an OTLP endpoint, see [Install an OTLP-based metrics collector](https://github.com/kyma-project/examples/tree/main/metrics-otlp).
- Learn how to [access and expose](../../../04-operation-guides/security/sec-06-access-expose-grafana.md) the services Grafana, Jaeger, and Kiali.

- Troubleshoot Observability-related issues:
  - [Prometheus Istio Server keeps crashing](../../../04-operation-guides/troubleshooting/observability/obsv-01-troubleshoot-prometheus-istio-server-crash-oom.md)
  - [Trace backend shows fewer traces than expected](../../../04-operation-guides/troubleshooting/observability/obsv-02-troubleshoot-trace-backend-shows-few-traces.md)
  - [Loki shows fewer logs than expected](../../../04-operation-guides/troubleshooting/observability/obsv-03-troubleshoot-loki-logging.md)

- Understand the architecture of Kyma's [monitoring](../../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring.md), [logging](../../../05-technical-reference/00-architecture/obsv-02-architecture-logging.md), and [tracing](../../../01-overview/main-areas/telemetry/telemetry-03-traces.md) components.

- Find the [configuration parameters for Monitoring, Logging, Tracing, and Kiali](../../../05-technical-reference/00-configuration-parameters/obsv-01-configpara-observability.md).

- [Deploy Kiali](https://github.com/kyma-project/examples/blob/main/kiali/README.md) to a Kyma cluster
