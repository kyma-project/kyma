---
title: Enable Istio access logs
---

You can enable [Istio access logs](https://istio.io/latest/docs/tasks/observability/logs/access-log/) to provide fine-grained details about the access to workloads that are part of the Istio service mesh. This can help in indicating the four “golden signals” of monitoring (latency, traffic, errors, and saturation), and troubleshooting anomalies.
The Istio setup shipped with Kyma provides a pre-configured [extension provider](https://istio.io/latest/docs/tasks/observability/telemetry) for access logs which will configure the istio-proxies to print access logs to stdout using JSON format. It uses a configuration like this:
```yaml
  extensionProviders:
    - name: stdout-json
      envoyFileAccessLog:
        path: "/dev/stdout"
        logFormat:
          labels:
            # default Istio log format plus relevant entries for trace context
            ...
            traceparent: "%REQ(TRACEPARENT)%"
            tracestate: "%REQ(TRACESTATE)%"

```
The [log format](https://github.com/kyma-project/kyma/blob/main/resources/istio/values.yaml#L62) is based on the Istio default format enhanced with the attributes relevant for identifying the related trace context conform to the [w3c-tracecontext](https://www.w3.org/TR/trace-context/) protocol. See [Kyma tracing](./../../01-overview/telemetry/telemetry-03-traces.md) for more details on tracing. See [Istio tracing](./../../01-overview/telemetry/telemetry-03-traces.md#istio) on how to enable trace context propagation with Istio.

>**CAUTION:** Enabling access logs may drastically increase logs volume and might quickly fill up your log storage. Also, the provided feature uses an API in alpha state, which may change in future releases.

## Configuration

Istio access logs can be enabled selectively using the Telemetry API. User can enable access logs for the entire Namespace, for a selective workload, or on Istio gateway scope.

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
  accessLogging:
    - providers:
      - name: stdout-json
```

### Configure Istio access logs for a specific gateway

Instead of enabling the access logs for all the individual proxies of the workloads you have, you can enable the logs for the proxy used by the related Istio ingress gateway:

```yaml
apiVersion: telemetry.istio.io/v1alpha1
kind: Telemetry
metadata:
  name: access-config
  namespace: istio-system
spec:
  selector:
    matchLabels:
      istio: ingressgateway
  accessLogging:
    - providers:
      - name: stdout-json
```