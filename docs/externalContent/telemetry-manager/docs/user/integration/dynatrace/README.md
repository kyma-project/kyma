# Integrate With Dynatrace

## Overview

| Category| |
| - | - |
| Signal types | traces, metrics |
| Backend type | third-party remote |
| OTLP-native | yes, but dynatrace agent in parallel |

[Dynatrace](https://www.dynatrace.com) is an advanced Application Performance Management solution available as SaaS offering. It supports monitoring both the Kubernetes cluster itself and the workloads running in the cluster. To use all of its features, the proprietary agent technology of Dynatrace must be installed.

With the Kyma Telemetry module, you gain even more visibility by adding custom spans and Istio spans, as well as custom metrics. Get an introduction on how to set up Dynatrace and learn how to integrate the Kyma Telemetry module.

![setup](./../assets/dynatrace.drawio.svg)

## Table of Content

- [Integrate With Dynatrace](#integrate-with-dynatrace)
  - [Overview](#overview)
  - [Table of Content](#table-of-content)
  - [Prerequisites](#prerequisites)
  - [Prepare the Namespace](#prepare-the-namespace)
  - [Dynatrace Setup](#dynatrace-setup)
  - [Telemetry Module Setup](#telemetry-module-setup)
    - [Create Secret](#create-secret)
    - [Ingest Traces](#ingest-traces)
    - [Ingest Metrics](#ingest-metrics)
  - [Set Up Kyma Dashboard Integration](#set-up-kyma-dashboard-integration)
  - [Use Dynatrace Dashboards](#use-dynatrace-dashboards)
  - [Use Dynatrace Alerts](#use-dynatrace-alerts)

## Prerequisites

- Kyma as the target deployment environment
- The [Telemetry module](https://kyma-project.io/#/telemetry-manager/user/README) is [added](https://kyma-project.io/#/02-get-started/01-quick-install)
- Active Dynatrace environment with permissions to create new access tokens
- Helm 3.x if you want to deploy the [OpenTelemetry sample application](../opentelemetry-demo/README.md)

## Prepare the Namespace

1. Export the namespace you want to use for Dynatrace as a variable with the following command:

    ```bash
    export DYNATRACE_NS="dynatrace"
    ```

1. If you haven't created a namespace yet, do it now:

    ```bash
    kubectl create namespace $DYNATRACE_NS
    ```

## Dynatrace Setup

There are different ways to deploy Dynatrace on Kubernetes. All [deployment options](https://docs.dynatrace.com/docs/ingest-from/setup-on-k8s/deployment) are based on the [Dynatrace Operator](https://github.com/Dynatrace/dynatrace-operator).

1. Install Dynatrace with the namespace you prepared earlier.
   > [!NOTE]
   > By default, Dynatrace used the classic full-stack injection. However, for better stability, we recommend using the [cloud-native fullstack injection](https://docs.dynatrace.com/docs/ingest-from/setup-on-k8s/guides/operation/migration/classic-to-cloud-native).

2. In the DynaKube resource, configure the correct `apiurl` of your environment.

3. In the DynaKube resource, exclude Kyma system namespaces by adding the following snippet:

    ```yaml
    spec:
      metadataEnrichment:
        enabled: true
        namespaceSelector:
          matchExpressions:
          - key: operator.kyma-project.io/managed-by
            operator: NotIn
            values:
              - kyma
      oneAgent:
        cloudNativeFullStack:
          autoUpdate: true
          namespaceSelector:
            matchExpressions:
            - key: operator.kyma-project.io/managed-by
              operator: NotIn
              values:
                - kyma
    ```

4. In the environment, go to **Settings > Cloud and virtualization > Kubernetes** and enable the relevant Kubernetes features, especially `Monitor annotated Prometheus exporters` to collect Istio metrics.

5. In the Dynatrace Hub, enable the **Istio Service Mesh** extension and annotate your services as outlined in the description.

6. If you have a workload exposing metrics in the Prometheus format, you can collect custom metrics in Prometheus format by [annotating the workload](https://docs.dynatrace.com/docs/observe/infrastructure-monitoring/container-platform-monitoring/kubernetes-monitoring/monitor-prometheus-metrics#annotate-kubernetes-services). If the workload has an Istio sidecar, you must either weaken the mTLS setting for the metrics port by defining an [Istio PeerAuthentication](https://istio.io/latest/docs/reference/config/security/peer_authentication/#PeerAuthentication) or exclude the port from interception by the Istio proxy by placing an `traffic.sidecar.istio.io/excludeInboundPorts` annotaion on your Pod that lists the metrics port.

As a result, you see data arriving in your environment, advanced Kubernetes monitoring is possible, and Istio metrics are available.

## Telemetry Module Setup

Next, you set up the ingestion of custom span and Istio span data, and, optionally, custom metrics based on OTLP.

### Create Secret

1. To push custom metrics and spans to Dynatrace, set up a [dataIngestToken](https://docs.dynatrace.com/docs/manage/identity-access-management/access-tokens-and-oauth-clients/access-tokens/personal-access-token).

   Follow the instructions in [Dynatrace: Generate an access token](https://docs.dynatrace.com/docs/manage/identity-access-management/access-tokens-and-oauth-clients/access-tokens/personal-access-token#generate-personal-access-tokens) and select the following scopes:

   - **Ingest metrics**
   - **Ingest OpenTelemetry traces**

2. Create an [apiToken](https://docs.dynatrace.com/docs/manage/identity-access-management/access-tokens-and-oauth-clients/access-tokens/personal-access-token) by selecting the template `Kubernetes: Dynatrace Operator`.

3. To create a new Secret containing your access tokens, replace the `<API_TOKEN>` and `<DATA_INGEST_TOKEN>` placeholder with the `apiToken` and `dataIngestToken` you created, replace the `<API_URL>` placeholder with the Dynatrace endpoint, and run the following command:

   ```bash
   kubectl -n $DYNATRACE_NS create secret generic dynakube --from-literal="apiToken=<API_TOKEN>" --from-literal="dataIngestToken=<DATA_INGEST_TOKEN>" --from-literal="apiurl=<API_URL>"
   ```

4. Verify the Secret you created looks similar to the [example Secret](https://github.com/kyma-project/telemetry-manager/blob/main/docs/user/integration/dynatrace/secret-example.yaml).

### Ingest Traces

To start ingesting custom spans and Istio spans, you must enable the Istio tracing feature and then deploy a TracePipeline.

1. Deploy the Istio Telemetry resource:

    ```bash
    kubectl apply -n istio-system -f - <<EOF
    apiVersion: telemetry.istio.io/v1
    kind: Telemetry
    metadata:
      name: tracing-default
    spec:
      tracing:
      - providers:
        - name: "kyma-traces"
        randomSamplingPercentage: 1.0
    EOF
    ```

    The default configuration has the **randomSamplingPercentage** property set to `1.0`, meaning it samples 1% of all requests. To change the sampling rate, adjust the property to the desired value, up to 100 percent.

    > [!WARNING]
    > Be cautious when you configure the **randomSamplingPercentage**:
    > - Could cause high volume of traces.
    > - The Kyma trace gateway component does not scale automatically.

1. Deploy the TracePipeline:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: telemetry.kyma-project.io/v1alpha1
    kind: TracePipeline
    metadata:
        name: dynatrace
    spec:
        output:
            otlp:
                endpoint:
                    valueFrom:
                        secretKeyRef:
                            name: dynakube
                            namespace: ${DYNATRACE_NS}
                            key: apiurl
                path: v2/otlp/v1/traces
                headers:
                    - name: Authorization
                      prefix: Api-Token
                      valueFrom:
                          secretKeyRef:
                              name: dynakube
                              namespace: ${DYNATRACE_NS}
                              key: dataIngestToken
                protocol: http
    EOF
    ```

1. To find traces from your Kyma cluster in the Dynatrace UI, go to **Applications & Microservices** > **Distributed traces**.

### Ingest Metrics

> [!NOTE]
> The Dynatrace OTLP API exclusively supports "delta" [aggregation temporality](https://docs.dynatrace.com/docs/ingest-from/opentelemetry/getting-started/metrics/limitations#aggregation-temporality). If your metrics are emitted in cumulative temporality (the default temporality in OTel), you must set up the transformation with a custom OTel Collector Deployment in addition to your MetricPipeline.

Depending on they way your applications emit metrics, choose one of the following approaches to ingest custom metrics to Dynatrace:

- Use a MetricPipeline together with a custom OTel Collector Deployment.

  This approach adds the required "cumulative to delta" transformation by running an additional custom OTel Collector. The Telemetry Metric gateway ships the metrics to the custom collector, and the collector transforms them before shipping the data to the Dynatrace endpoint. This approach enables support for all metric types and inputs for the MetricPipeline. However, you must operate the additional OTel Collector in a custom way.

  1. Deploy the custom OTel Collector using Helm with a custom [values.yaml](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/dynatrace/exporter-values.yaml):

        ```bash
        helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
        helm repo update

        helm upgrade --install -n ${DYNATRACE_NS} dynatrace-exporter open-telemetry/opentelemetry-collector -f https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/dynatrace/exporter-values.yaml
        ```

  1. Deploy the MetricPipeline that ships to the custom OTel Collector:

        ```bash
        cat <<EOF | kubectl apply -f -
        apiVersion: telemetry.kyma-project.io/v1alpha1
        kind: MetricPipeline
        metadata:
            name: dynatrace
        spec:
            input:
                istio:
                    enabled: true
                prometheus:
                    enabled: true
            output:
                otlp:
                    endpoint:
                        value: http://dynatrace-exporter-opentelemetry-collector.${DYNATRACE_NS}:4317
        EOF
        ```

- If your application pushes metrics in delta temporality with OTLP, you can use a MetricPipeline to push metrics directly.

  To use this setup, you must explicitly enable the "delta" aggregation temporality in your applications. You cannot enable additional inputs for the MetricPipeline because these produce metrics with "cumulative" temperability.

  1. Deploy the MetricPipeline:

        ```bash
        cat <<EOF | kubectl apply -f -
        apiVersion: telemetry.kyma-project.io/v1alpha1
        kind: MetricPipeline
        metadata:
            name: dynatrace
        spec:
            output:
                otlp:
                    endpoint:
                        valueFrom:
                            secretKeyRef:
                                name: dynakube
                                namespace: ${DYNATRACE_NS}
                                key: apiurl
                    path: v2/otlp/v1/metrics
                    headers:
                        - name: Authorization
                          prefix: Api-Token
                          valueFrom:
                            secretKeyRef:
                              name: dynakube
                              namespace: ${DYNATRACE_NS}
                              key: dataIngestToken
                    protocol: http
        EOF
        ```

  1. Start pushing metrics to the metric gateway using [delta aggregation temporality.](https://docs.dynatrace.com/docs/ingest-from/opentelemetry/getting-started/metrics/limitations#aggregation-temporality)

  1. To find metrics from your Kyma cluster in the Dynatrace UI, go to **Observe & Explore** > **Metrics**.

- If your workloads expose metrics in the typical Prometheus format, use the native Dynatrace metric ingestion with Prometheus exporters:

  Use the [Dynatrace annotation approach](https://docs.dynatrace.com/docs/observe/infrastructure-monitoring/container-platform-monitoring/kubernetes-monitoring/monitor-prometheus-metrics), where the Dynatrace ActiveGate component running in your cluster scrapes workloads that are annotated with Dynatrace-specific annotations.

  This approach works well with workloads that expose metrics in the typical Prometheus format when not running with Istio.
  If you use Istio, you must disable Istio interception for the relevant metric port with the [traffic.istio.io/excludeInboundPorts](https://istio.io/latest/docs/reference/config/annotations/#TrafficExcludeInboundPorts) annotation. To collect Istio metrics from the envoys themselves, you need additional Dynatrace annotations for every workload.

## Set Up Kyma Dashboard Integration

For easier access from the Kyma dashboard, add links to new navigation under **Dynatrace**.

1. Apply the ConfigMaps:

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/dynatrace/kyma-dashboard-configmap.yaml
    ```

2. If your Secret has a different name or namespace, then download the file first and adjust the namespace and name accordingly in the 'dataSources' section of the file.

## Use Dynatrace Dashboards

1. To see the health of the Kyma Telemetry module and its related pipelines, import the file [Telemetry Module Status](./telemetry-resource-metrics.json) as a Dynatrace dashboard. For details, see [Importing Dashboards](https://docs.dynatrace.com/docs/analyze-explore-automate/dashboards-classic/dashboards/dashboard-json#import-dashboard).

2. Add the following custom resource attributes to the allow list of OpenTelemetry metrics resource attributes:
   - `k8s.resource.name`
   - `k8s.resource.group`
   - `k8s.resource.kind`
   - `k8s.resource.version`

   For details about adding attributes to the allow list, see [Configure resource and scope attributes to be added as dimensions](https://docs.dynatrace.com/docs/ingest-from/opentelemetry/getting-started/metrics/configuration#allow-list).

## Use Dynatrace Alerts

To send alerts about the Kyma Telemetry module status to your preferred backend system, create Dynatrace alerts based on certain metric events:

1. To define how and when alerts are triggered, create a problem alerting profile. For details, see [Create an alerting profile](https://docs.dynatrace.com/docs/analyze-explore-automate/notifications-and-alerting/alerting-profiles#create-an-alerting-profile).
2. To push alerts to your backend system, set up problem notifications in Dynatrace. For details, see [Problem notifications](https://docs.dynatrace.com/docs/analyze-explore-automate/notifications-and-alerting/problem-notifications).
3. Create a metric event with a metric selector or a metric key that reflects the event you want to monitor. For details, see [Metric events](https://docs.dynatrace.com/docs/discover-dynatrace/platform/davis-ai/anomaly-detection/set-up-a-customized-anomaly-detector/how-to-set-up/metric-events).
   For example, trigger an alert when the Kyma Telemetry module enters a non-ready state:
     ```text
     kyma.resource.status.state:filter(not(eq("state","Ready")))
     ```
4. To target the metric event you just created, add a custom event filter in your alerting profile. For details, see [event filters](https://docs.dynatrace.com/docs/analyze-explore-automate/notifications-and-alerting/alerting-profiles#event-filters).
5. To test the integration, trigger the metric event and confirm that the target system receives the alert.

