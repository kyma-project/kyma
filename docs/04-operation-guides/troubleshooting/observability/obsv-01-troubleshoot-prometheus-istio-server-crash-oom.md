---
title: Prometheus Istio Server restarting or in crashback loop
---

> **NOTE:** Prometheus and Grafana are [deprecated](https://github.com/kyma-project/website/blob/main/content/blog-posts/2022-12-09-monitoring-deprecation/index.md) and are planned to be removed. If you want to install a custom stack, take a look at [Install a custom kube-prometheus-stack in Kyma](https://github.com/kyma-project/examples/tree/main/prometheus).

## Condition

Prometheus Istio Server is restarting or in a crashback loop.

## Cause

Prometheus Istio Server scrapes metrics from all envoy side cars, which may lead to OOM issues.

For example, this can happen when a high number of workloads perform a lot of communication to other workloads, or when workloads are created and deleted dynamically.

In such cases, the cardinality of the Istio metrics may increase too much and cause the container to be killed because of OOM (Istio telemetry V2 currently doesn't support the concept of metric expiry).

There can be other causes for the Prometheus Istio Server to restart or crash, but the following istructions focus on fixing the OOM issue.

## Remedy

To prevent the OOM issue, you can increase the memory limit.
Additionally, you can choose to decrease the volume of data by dropping additional labels.

> **CAUTION:** Dropping additional labels with `prometheus-istio.envoyStats.labeldropRegex` has the side effect that graphs in Kiali don't work.

For both solutions, you can choose to change your Kyma cluster settings or directly update the Istio Prometheus resources.

### Change the Kyma settings

1. To increase the memory limit, create a values YAML file with the following content:

   ```yaml
   monitoring:
     prometheus-istio:
       server:
         resources:
           limits:
             memory: "6Gi"
   ```
  
   > **TIP:** You should be fine with increasing the limit to 6Gi. However, if your resources are scarce, try increasing the value gradually in steps of 1Gi.

2. Deploy the values YAML file with the following command:

   ```bash
   kyma deploy --values-file {VALUES_FILE_PATH}
   ```

3. If the problem persists, drop additional labels for the Istio metrics with the following values YAML file:
  
   ```yaml
   monitoring:
     prometheus-istio:
       envoyStats:
         labeldropRegex: "^(grpc_response_status|source_version|source_principal|source_app|response_flags|request_protocol|destination_version|destination_principal|destination_app|destination_canonical_service|destination_canonical_revision|source_canonical_revision|source_canonical_service)$"
   ```

4. Change the settings with the following command:

   ```bash
   kyma deploy --values-file {VALUES_FILE_PATH}
   ```

### Change the Istio Prometheus configuration

1. To increase the memory for `prometheus-istio-server`, run the following command:
  
   ```bash
   kubectl edit deployment -n kyma monitoring-prometheus-istio-server
   ```

2. In your deployment resource, set the following limits for memory:

   ```yaml
   resources:
     limits:
       cpu: 600m
       memory: 6000Mi
     requests:
       cpu: 40m
       memory: 200Mi
   ```
  
   > **TIP:** You should be fine with increasing the limit to 6Gi. However, if your resources are scarce, try increasing the value gradually in steps of 1Gi.

3. If the problem persists, drop additional labels for the Istio metrics by editing `prometheus-istio server`:

   ```bash
   kubectl edit configmap -n kyma-system monitoring-prometheus-istio-server
   ```

4. Modify the following values:

   ```yaml
   metric_relabel_configs:
     - separator: ;
       regex: ^(grpc_response_status|source_version|destination_version|source_app|destination_app)$
       replacement: $1
       action: labeldrop
   ```

5. Change regex in the following way:

   ```yaml
   regex: ^(grpc_response_status|source_version|source_principal|source_app|response_flags|request_protocol|destination_version|destination_principal|destination_app|destination_canonical_service|destination_canonical_revision|source_canonical_revision|source_canonical_service)$
   ```

6. Save the ConfigMap and restart `prometheus-istio-server` for the changes to take effect:

   ```bash
   kubectl rollout restart deployment -n kyma-system monitoring-prometheus-istio-server
   ```
