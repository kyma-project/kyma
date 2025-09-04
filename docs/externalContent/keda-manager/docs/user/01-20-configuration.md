# Configuring Keda Module

By default, the Keda module comes with the default configuration. You can change the configuration using the Keda CustomResourceDefinition (CRD). See how to configure the **logging.level** attribute, enable the Istio sidecar injection, change resource consumption, define custom annotations, or override the minimum TLS version.

## Prerequisites

[You have added the Keda module](https://kyma-project.io/#/02-get-started/01-quick-install).

## Procedure

1. Go to Kyma dashboard.
2. Choose **Modify Modules**, and in the **View** tab, choose `keda`.
3. Go to **Edit**, and provide your configuration changes. You can use the **Form** or **YAML** tab.

- To define the level of detail of your logs, set the **logging.level** attribute to one of the following values:
   - `debug` - is the most detailed option. Useful for a developer during debugging.
   - `info` - provides standard log level indicating operations within the Keda module. For example, it can show whether the workload scaling operation was successful or not.
   - `error` - shows error logs only. This means only log messages corresponding to errors and misconfigurations are visible in logs.

   ```yaml
   spec:
     logging:
       operator:
         level: "debug"
   ```

- To enable the Istio sidecar injection for **operator** and **metricServer**, set the value of **enabledSidecarInjection** to `true`. For example:

  ```yaml
  spec:
    istio:
      metricServer:
        enabledSidecarInjection: true
      operator:
        enabledSidecarInjection: true
  ```

- To change the resource consumption, enter your preferred values for **operator**, **metricServer** and **admissionWebhook**. For example:

   ```yaml
   spec:
     resources:
       operator:
         limits:
           cpu: "1"
           memory: "200Mi"
         requests:
           cpu: "150m"
           memory: "150Mi"
       metricServer:
         limits:
           cpu: "1"
           memory: "1000Mi"
         requests:
           cpu: "150m"
           memory: "500Mi"
       admissionWebhook:
         limits:
           cpu: "1"
           memory: "1000Mi"
         requests:
           cpu: "50m"
           memory: "800Mi"
   
   ```

- To define custom annotations for KEDA workloads, enter your preferred values for **operator**, **metricServer** and **admissionWebhook**. For example:

   ```yaml
   spec:
     podAnnotations:
      operator:
        metrics.dynatrace.com/scrape: 'true'
        metrics.dynatrace.com/path: '/metrics'
      metricServer:
        metrics.dynatrace.com/scrape: 'true'
        metrics.dynatrace.com/path: '/metrics'
      admissionWebhook:
        metrics.dynatrace.com/scrape: 'true'
        metrics.dynatrace.com/path: '/metrics'
   
   ```

- To override the minimum TLS version used by KEDA (default is `TLS12`), set the `KEDA_HTTP_MIN_TLS_VERSION` environment variable. For example:

   ```yaml
   spec:
     env:
       - name: KEDA_HTTP_MIN_TLS_VERSION
         value: TLS13
   ```

For more information about the Keda resources, see [Keda concepts](https://keda.sh/docs/latest/concepts/).
