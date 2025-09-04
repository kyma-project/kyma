# Integrate with SAP Cloud Logging

| Category     |                                         |
| ------------ | --------------------------------------- |
| Signal types | logs, traces, metrics                   |
| Backend type | third-party remote                      |
| OTLP-native  | yes for traces and metrics, no for logs |

Learn how to configure the Telemetry module to ingest application and access logs as well as distributed trace data and metrics in instances of SAP Cloud Logging.

## Table of Content

- [Integrate with SAP Cloud Logging](#integrate-with-sap-cloud-logging)
  - [Table of Content](#table-of-content)
  - [Prerequisites](#prerequisites)
  - [Context](#context)
  - [Ship Logs to SAP Cloud Logging](#ship-logs-to-sap-cloud-logging)
    - [Set Up Application Logs](#set-up-application-logs)
    - [Set Up Access Logs](#set-up-access-logs)
  - [Ship Distributed Traces to SAP Cloud Logging](#ship-distributed-traces-to-sap-cloud-logging)
  - [Ship Metrics to SAP Cloud Logging](#ship-metrics-to-sap-cloud-logging)
  - [Set Up Kyma Dashboard Integration](#set-up-kyma-dashboard-integration)
  - [Use SAP Cloud Logging Alerts](#use-sap-cloud-logging-alerts)
  - [Use SAP Cloud Logging Dashboards](#use-sap-cloud-logging-dashboards)

## Prerequisites

- Kyma as the target deployment environment.
- The [Telemetry module](../../README.md) is added. For details, see [Quick Install](https://kyma-project.io/#/02-get-started/01-quick-install). <!-- This link differs for OS and SKR -->
- If you want to use Istio access logs, make sure that the [Istio module](https://kyma-project.io/#/istio/user/README) is added.
- An instance of [SAP Cloud Logging](https://help.sap.com/docs/cloud-logging?locale=en-US&version=Cloud) with OpenTelemetry enabled to ingest distributed traces.
  > [!TIP]
  > Create the instance with the SAP BTP service operator (see [Create an SAP Cloud Logging Instance through SAP BTP Service Operator](https://help.sap.com/docs/cloud-logging/cloud-logging/create-sap-cloud-logging-instance-through-sap-btp-service-operator?locale=en-US&version=Cloud)), because it takes care of creation and rotation of the required Secret. However, you can choose any other method of creating the instance and the Secret, as long as the parameter for OTLP ingestion is enabled in the instance. For details, see [Configuration Parameters](https://help.sap.com/docs/cloud-logging/cloud-logging/configuration-parameters?locale=en-US&version=Cloud).
- A Secret in the respective namespace in the Kyma cluster, holding the credentials and endpoints for the instance. In the following example, the Secret is named `sap-cloud-logging` and the namespace `sap-cloud-logging-integration`, as illustrated in the [secret-example.yaml](https://github.com/kyma-project/telemetry-manager/blob/main/docs/user/integration/sap-cloud-logging/secret-example.yaml).
<!-- markdown-link-check-disable -->
- Kubernetes CLI (kubectl) (see [Install the Kubernetes Command Line Tool](https://developers.sap.com/tutorials/cp-kyma-download-cli.html)).
<!-- markdown-link-check-enable -->
- UNIX shell or Windows Subsystem for Linux (WSL) to execute commands.

## Context

The Telemetry module supports shipping logs and ingesting distributed traces as well as metrics from applications and the Istio service mesh to SAP Cloud Logging. Furthermore, you can set up Kyma dashboard integration and use SAP Cloud Logging alerts and dashboards.

SAP Cloud Logging is an instance-based and environment-agnostic observability service to store, visualize, and analyze logs, metrics, and traces.

![setup](./../assets/sap-cloud-logging.drawio.svg)

## Ship Logs to SAP Cloud Logging

You can set up shipment of applications and access logs to SAP Cloud Logging. The following instructions distinguish application logs and access logs, which can be configured independently.

### Set Up Application Logs
<!-- using HTML so it's collapsed in GitHub, don't switch to docsify tabs -->
1. Deploy the LogPipeline for application logs with the following script:

   <div tabs name="applicationlogs">
     <details><summary>Script: Application Logs</summary>

    ```bash
    kubectl apply -n sap-cloud-logging-integration -f - <<EOF
    apiVersion: telemetry.kyma-project.io/v1alpha1
    kind: LogPipeline
    metadata:
      name: sap-cloud-logging-application-logs
    spec:
      input:
        application:
          containers:
            exclude:
              - istio-proxy
      output:
        http:
          dedot: true
          host:
            valueFrom:
              secretKeyRef:
                name: sap-cloud-logging
                namespace: sap-cloud-logging-integration
                key: ingest-mtls-endpoint
          tls:
            cert:
              valueFrom:
                secretKeyRef:
                  name: sap-cloud-logging
                  namespace: sap-cloud-logging-integration
                  key: ingest-mtls-cert
            key:
              valueFrom:
                secretKeyRef:
                  name: sap-cloud-logging
                  namespace: sap-cloud-logging-integration
                  key: ingest-mtls-key
          uri: /customindex/kyma
    EOF
    ```

      </details>
    </div>

2. Wait for the LogPipeline to be in the `Running` state. To check the state, run: `kubectl get logpipelines`.

### Set Up Access Logs

By default, Istio sidecar injection and Istio access logs are disabled in Kyma. To analyze access logs of your workload in the default SAP Cloud Logging dashboards shipped for SAP BTP, Kyma runtime, you must enable them:

1. Enable Istio sidecar injection for your workload. See [Enabling Istio Sidecar Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection).

2. Enable Istio access logs for your workload. See [Configure Istio Access Logs](https://kyma-project.io/#/istio/user/tutorials/01-45-enable-istio-access-logs).

3. Deploy the LogPipeline for Istio access logs and enable access logs in Kyma with the following script:

   <div tabs name="accesslogs">
     <details><summary>Script: Access Logs</summary>

    ```bash
    kubectl apply -n sap-cloud-logging-integration -f - <<EOF
    apiVersion: telemetry.kyma-project.io/v1alpha1
    kind: LogPipeline
    metadata:
      name: sap-cloud-logging-access-logs
    spec:
      input:
        application:
          containers:
            include:
              - istio-proxy
      output:
        http:
          dedot: true
          host:
            valueFrom:
              secretKeyRef:
                name: sap-cloud-logging
                namespace: sap-cloud-logging-integration
                key: ingest-mtls-endpoint
          tls:
            cert:
              valueFrom:
                secretKeyRef:
                  name: sap-cloud-logging
                  namespace: sap-cloud-logging-integration
                  key: ingest-mtls-cert
            key:
              valueFrom:
                secretKeyRef:
                  name: sap-cloud-logging
                  namespace: sap-cloud-logging-integration
                  key: ingest-mtls-key
          uri: /customindex/istio-envoy-kyma
    EOF
    ```

      </details>
    </div>

4. Wait for the LogPipeline to be in the `Running` state. To check the state, run: `kubectl get logpipelines`.

## Ship Distributed Traces to SAP Cloud Logging

You can set up ingestion of distributed traces from applications and the Istio service mesh to the OTLP endpoint of the SAP Cloud Logging service instance.

1. Deploy the Istio Telemetry resource with the following script:

   <div tabs name="istiotraces">
     <details><summary>Script: Istio Traces</summary>

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

     </details>
   </div>

   The default configuration has the **randomSamplingPercentage** property set to `1.0`, meaning it samples 1% of all requests. To change the sampling rate, adjust the property to the desired value, up to 100 percent.

2. Deploy the TracePipeline with the following script:

   <div tabs name="distributedtraces">
     <details><summary>Script: Distributed Traces</summary>

    ```bash
    kubectl apply -n sap-cloud-logging-integration -f - <<EOF
    apiVersion: telemetry.kyma-project.io/v1alpha1
    kind: TracePipeline
    metadata:
      name: sap-cloud-logging
    spec:
      output:
        otlp:
          endpoint:
            valueFrom:
              secretKeyRef:
                name: sap-cloud-logging
                namespace: sap-cloud-logging-integration
                key: ingest-otlp-endpoint
          tls:
            cert:
              valueFrom:
                secretKeyRef:
                  name: sap-cloud-logging
                  namespace: sap-cloud-logging-integration
                  key: ingest-otlp-cert
            key:
              valueFrom:
                secretKeyRef:
                  name: sap-cloud-logging
                  namespace: sap-cloud-logging-integration
                  key: ingest-otlp-key
    EOF
    ```

     </details>
   </div>

3. Wait for the TracePipeline to be in the `Running` state. To check the state, run: `kubectl get tracepipelines`.

## Ship Metrics to SAP Cloud Logging

You can set up ingestion of metrics from applications and the Istio service mesh to the OTLP endpoint of the SAP Cloud Logging service instance.

1. Deploy the MetricPipeline with the following script:

   <div tabs name="SAPCloudLogging">
     <details><summary>Script: SAP Cloud Logging</summary>

    ```bash
    kubectl apply -n sap-cloud-logging-integration -f - <<EOF
    apiVersion: telemetry.kyma-project.io/v1alpha1
    kind: MetricPipeline
    metadata:
      name: sap-cloud-logging
    spec:
      input:
        prometheus:
          enabled: false
        istio:
          enabled: false
        runtime:
          enabled: false
      output:
        otlp:
          endpoint:
            valueFrom:
              secretKeyRef:
                name: sap-cloud-logging
                namespace: sap-cloud-logging-integration
                key: ingest-otlp-endpoint
          tls:
            cert:
              valueFrom:
                secretKeyRef:
                  name: sap-cloud-logging
                  namespace: sap-cloud-logging-integration
                  key: ingest-otlp-cert
            key:
              valueFrom:
                secretKeyRef:
                  name: sap-cloud-logging
                  namespace: sap-cloud-logging-integration
                  key: ingest-otlp-key
    EOF
    ```

     </details>
   </div>

    By default, the MetricPipeline assures that a gateway is running in the cluster to push OTLP metrics.

2. If you want to use additional metric collection, configure the presets under `input`.

   For the available options, see [Metrics](./../../04-metrics.md).

3. Wait for the MetricPipeline to be in the `Running` state. To check the state, run: `kubectl get metricpipelines`.

## Set Up Kyma Dashboard Integration

For easier access from the Kyma dashboard, add links to the navigation under **SAP Cloud Logging**, and add deep links to the **Pod**, **Deployment**, and **Namespace** views.

Depending on the output you use in your LogPipeline, apply the ConfigMap. If your Secret has a different name or namespace, then download the file first and adjust the namespace and name accordingly in the 'dataSources' section of the file.

- For OTLP, run:

  ```bash
  kubectl apply -f https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/kyma-dashboard-configmap.yaml
  ```

- For HTTP, run:

  ```bash
  kubectl apply -f https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/kyma-dashboard-http-configmap.yaml
  ```

## Use SAP Cloud Logging Alerts

Learn how to define and import recommended alerts for SAP Cloud Logging. The following alerts are based on JSON documents defining a `Monitor` for the alerting plugin.

1. Define a `destination`, which will be used by all your alerts.
2. To import a monitor, use the development tools of the SAP Cloud Logging dashboard.
3. Execute `POST _plugins/_alerting/monitors`, followed by the contents of the respective JSON file.
4. Depending on the pipelines you are using, enable the some or all of the following alerts:
<!-- markdown-link-check-disable -->
   | Category                   | File                                                                                                                                                                    | Description                                                                                                                                                                                   |
   | -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
   | SAP Cloud Logging          | [OpenSearch cluster health](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/alert-health.json)            | The OpenSearch cluster might become unhealthy, which is indicated by a "red" status using the [cluster health API](https://opensearch.org/docs/1.3/api-reference/cluster-api/cluster-health). |
   | Kyma Telemetry Integration | [Application log ingestion](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/alert-app-log-ingestion.json) | The LogPipeline for shipping [application logs](#ship-logs-to-sap-cloud-logging) might lose connectivity to SAP Cloud Logging, with the effect that no application logs are ingested anymore. |
   | Kyma Telemetry Integration | [Access log ingestion](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/alert-access-log-ingestion.json)   | The LogPipeline for shipping [access logs](#ship-logs-to-sap-cloud-logging) might lose connectivity to SAP Cloud Logging, with the effect that no access logs are ingested anymore.           |
   | Kyma Telemetry Integration | [Trace ingestion](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/alert-trace-ingestion.json)             | The TracePipeline for shipping [traces](#ship-distributed-traces-to-sap-cloud-logging) might lose connectivity to SAP Cloud Logging, with the effect that no traces are ingested anymore.     |
   | Kyma Telemetry Integration | [Metric ingestion](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/alert-metric-ingestion.json)           | The MetricPipeline for shipping [metrics](#ship-metrics-to-sap-cloud-logging) might lose connectivity to SAP Cloud Logging, with the effect that no metrics are ingested anymore.             |
   | Kyma Telemetry Integration | [Kyma Telemetry Status](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/alert-telemetry-status.json)      | The Telemetry module might report a non-ready state indicating a configuration or data flow problem.                                                                                          |
<!-- markdown-link-check-enable -->
5. Edit notification action: Add the `destination` and adjust the intervals and thresholds to your needs.
6. Verify that the new monitor definition is listed among the SAP Cloud Logging alerts.

## Use SAP Cloud Logging Dashboards

You can view logs, traces, and metrics in SAP Cloud Logging dashboards:

<!-- markdown-link-check-disable -->
- To view the traffic and application logs, use the SAP Cloud Logging dashboards prefixed with `Kyma_`, which are based on both kinds of log ingestion: application and access logs.
- To view distributed traces, use the OpenSearch plugin **Observability**.
- To view the container- and Pod-related metrics collected by the MetricPipeline `runtime` input, use the dashboard **[OTel] K8s Container Metrics**.
- To view the Kubernetes Node-related metrics collected by the MetricPipeline `runtime` input, manually import the file [K8s Nodes](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/dashboard-nodes.ndjson).
- To view the Kubernetes Volume-related metrics collected by the MetricPipeline `runtime` input, manually import the file [K8s Volumes](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/dashboard-volumes.ndjson).
- To view the Kubernetes Workload-related metrics collected by the MetricPipeline `runtime` input, manually import the file [K8s Workloads](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/dashboard-workloads.ndjson).
- To view the status of the SAP Cloud Logging integration with the Kyma Telemetry module, manually import the file [Kyma Telemetry Status](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/dashboard-status.ndjson).
- To use the dashboard for Istio metrics of Pods that have an active Istio sidecar injection (collected by the MetricPipeline `istio` input), manually import the file [Kyma Istio Service Metrics](https://raw.githubusercontent.com/kyma-project/telemetry-manager/main/docs/user/integration/sap-cloud-logging/dashboard-istio.ndjson).
<!-- markdown-link-check-enable -->
