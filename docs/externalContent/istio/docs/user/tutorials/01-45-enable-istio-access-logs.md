# Configure Istio Access Logs
Use the Telemetry API to selectively enable the Istio access logs and filter them if needed.

## Prerequisites
* You have the Istio module added.
* To use CLI instruction, you must install [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) and [curl](https://curl.se/). Alternatively, you can use Kyma dashboard.

## Context

You can enable [Istio access logs](https://istio.io/latest/docs/tasks/observability/logs/access-log/) to provide fine-grained details about the access to workloads that are part of the Istio service mesh. This can help indicate the four “golden signals” of monitoring (latency, traffic, errors, and saturation) and troubleshooting anomalies.
The Istio setup shipped with the Istio module provides a pre-configured [extension provider](https://istio.io/latest/docs/tasks/observability/telemetry) for access logs, which configures the Istio proxies to print access logs to stdout using the JSON format. It uses a configuration similar to the following one:

```yaml
extensionProviders:
  - name: stdout-json
    envoyFileAccessLog:
      path: "/dev/stdout"
      logFormat:
        labels:
          ...
          traceparent: "%REQ(TRACEPARENT)%"
          tracestate: "%REQ(TRACESTATE)%"
```

The [log format](https://github.com/kyma-project/istio/blob/main/internal/istiooperator/istio-operator.yaml#L160) is based on the Istio default format enhanced with the attributes relevant for identifying the related trace context conform to the [w3c-tracecontext](https://www.w3.org/TR/trace-context/) protocol. See [Kyma tracing](https://kyma-project.io/#/telemetry-manager/user/03-traces) for more details on tracing. See [Istio tracing](https://kyma-project.io/#/telemetry-manager/user/03-traces?id=istio) on how to enable trace context propagation with Istio.

> [!WARNING]
>  Enabling access logs may drastically increase logs volume and might quickly fill up your log storage.

## Configuration

Use the Telemetry API to selectively enable Istio access logs. See:

<!-- no toc -->
- [Configure Istio Access Logs for a Namespace](#configure-istio-access-logs-for-a-namespace)
- [Configure Istio Access Logs for a Selective Workload](#configure-istio-access-logs-for-a-selective-workload)
- [Configure Istio Access Logs for a Specific Gateway](#configure-istio-access-logs-for-a-selective-gateway)
- [Configure Istio Access Logs for the Entire Mesh](#configure-istio-access-logs-for-the-entire-mesh)

To filter the enabled access logs, you can edit the Telemetry API by adding a filter expression. See [Filter Access logs](#filter-access-logs).

### Configure Istio Access Logs for a Namespace

<Tabs>
<Tab name="Kyma Dashboard">

1. Go to the namespace for which you want to configure Istio access logs.
2. Go to **Istio > Telemetries** and select **Create**.
3. Provide a name, for example, `access-config`.
4. Select **Create**.
</Tab>
<Tab name="kubectl">

1. Export the name of the namespace for which you want to configure Istio access logs.
    
    ```bash
    export YOUR_NAMESPACE={NAMESPACE_NAME}
    ```

2. To apply the configuration, run:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: telemetry.istio.io/v1
    kind: Telemetry
    metadata:
      name: access-config
      namespace: $YOUR_NAMESPACE
    spec:
      accessLogging:
        - providers:
          - name: stdout-json
    EOF
    ```
3. To verify that the resource is applied, run:
    ```bash
    kubectl -n $YOUR_NAMESPACE get telemetries.telemetry.istio.io
    ```
</Tab>
</Tabs>


### Configure Istio Access Logs for a Selective Workload

To configure label-based selection of workloads, use a [selector](https://istio.io/latest/docs/reference/config/type/workload-selector/#WorkloadSelector).

<Tabs>
<Tab name="Kyma Dashboard">

1. Go to the namespace of the workloads for which you want to configure Istio access logs.
2. Go to **Istio > Telemetries** and select **Create**.
3. Switch to the `YAML` section and paste the following sample configuration into the editor:
    ```yaml
    apiVersion: telemetry.istio.io/v1
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
4. Replace `{YOUR_LABEL}` with the workloads' label and `{YOUR_NAMESPACE}` with the name of the workloads' namespace.
5. Select **Create**.
</Tab>
<Tab name="kubectl">

1. Export the name of the workloads' namespace and their label as environment variables:
    
    ```bash
    export YOUR_NAMESPACE={NAMESPACE_NAME}
    export YOUR_LABEL={LABEL}
    ```

2. To apply the configuration, run:
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: telemetry.istio.io/v1
    kind: Telemetry
    metadata:
      name: access-config
      namespace: $YOUR_NAMESPACE
    spec:
      selector:
        matchLabels:
          service.istio.io/canonical-name: $YOUR_LABEL
      accessLogging:
        - providers:
          - name: stdout-json
    EOF
    ```
3. To verify that the resource is applied, run:
    ```bash
    kubectl -n $YOUR_NAMESPACE get telemetries.telemetry.istio.io
    ```
</Tab>
</Tabs>

### Configure Istio Access Logs for a Selective Gateway

Instead of enabling the access logs for all the individual proxies of the workloads you have, you can enable the logs for the proxy used by the related Istio Ingress Gateway.

<Tabs>
<Tab name="Kyma Dashboard">

1. Go to the `istio-system` namespace.
2. Go to **Istio > Telemetries**.
3. Select **Create**.
4. Switch to the `YAML` section and paste the following sample configuration into the editor:
    ```yaml
    apiVersion: telemetry.istio.io/v1
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
5. Select **Create**.
</Tab>
<Tab name="kubectl">

1. To apply the configuration, run:
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: telemetry.istio.io/v1
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
    EOF
    ```
2. To verify that the resource is applied, run:
    ```bash
    kubectl -n istio-system get telemetries.telemetry.istio.io
    ```
</Tab>
</Tabs>

### Configure Istio Access Logs for the Entire Mesh

Enable access logs for all individual proxies of the workloads and Istio Ingress Gateways.

<Tabs>
<Tab name="Kyma Dashboard">

1. Go to the `istio-system` namespace.
2. Go to **Istio > Telemetries** and select **Create**.
3. Switch to the `YAML` section and paste the following sample configuration into the editor:
    ```yaml
    apiVersion: telemetry.istio.io/v1
    kind: Telemetry
    metadata:
      name: access-config
      namespace: istio-system
    spec:
      accessLogging:
        - providers:
          - name: stdout-json
    ```
4. Select **Create**.
</Tab>
<Tab name="kubectl">

1. To apply the configuration, run:
    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: telemetry.istio.io/v1
    kind: Telemetry
    metadata:
      name: access-config
      namespace: istio-system
    spec:
      accessLogging:
        - providers:
          - name: stdout-json
    EOF
    ```
2. To verify that the resource is applied, run:
    ```bash
    kubectl -n istio-system get telemetries.telemetry.istio.io
    ```
</Tab>
</Tabs>

### Filter Access Logs

Often, access logs emitted by Envoy do not contain data relevant to your observations, especially when the traffic is not based on an HTTP-based protocol. In such a situation, you can directly configure the Istio Envoys to filter out logs using a filter expression. To filter access logs, you can leverage the same [Istio Telemetry API](https://istio.io/latest/docs/reference/config/telemetry/#AccessLogging) that you used to enable them. To formulate which logs to **keep**, define a filter expression leveraging the typical [Envoy attributes](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/attributes).

For example, to filter out all logs having no protocol defined (which is the case if they are not HTTP-based), you can use a configuration similar to this example:
```yaml
apiVersion: telemetry.istio.io/v1
kind: Telemetry
metadata:
 name: access-config
 namespace: istio-system
spec:
 accessLogging:
 - filter:
     expression: 'has(request.protocol)'
   providers:
   - name: stdout-json
```
