---
title: Enable Istio access logs
---

You can enable [Istio access logs](https://istio.io/latest/docs/tasks/observability/logs/access-log/) to provide fine-grained details about the workloads. This can help in indicating the four “golden signals” of monitoring (latency, traffic, errors, and saturation), and troubleshooting anomalies.


>**CAUTION:** Enabling access logs may drastically increase logs volume and might quickly fill up your log storage. Also, the provided feature uses an API in alpha state, which may or may not be continued in future releases.


## Configuration

Istio access logs can be enabled selectively using the Telemetry API. User can enable access logs for the entire namespace or for a selective workload.

### Configure Istio access logs for the entire Namespace

1. In the following sample configuration, replace `{YOUR_NAMESPACE}` with your Namespace.
2. To apply the configuration, run `kubectl apply`.

```yaml
apiVersion: telemetry.istio.io/v1alpha1
kind: Telemetry
metadata:
  name: access-config
  namespace: {YOUR_NAMESPACE}
spec:
  accessLogging:
    - providers:
      - name: stdout-json
```

### Configure Istio access logs for a selective workload

To configure label-based selection of workloads, use a [selector](https://istio.io/latest/docs/reference/config/type/workload-selector/#WorkloadSelector).
1. In the following sample configuration, replace `{YOUR_NAMESPACE}` and `{YOUR_LABEL}` with your Namespace and the label of the workload, respectively.
2. To apply the configuration, run `kubectl apply`.

```yaml
apiVersion: telemetry.istio.io/v1alpha1
kind: Telemetry
metadata:
  name: access-config
  namespace: {YOUR_NAMESPACE}
spec:
  selector:
    matchLabels:
      service.istio.io/canonical-name: {YOUR_LABEL}
```