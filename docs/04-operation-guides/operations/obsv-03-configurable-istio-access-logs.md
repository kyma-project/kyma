---
title: Configure Istio access logs for Cloud Logging Service
---

[Istio access logs](https://istio.io/latest/docs/tasks/observability/logs/access-log/) can be enabled on a selective basis to provide fine-grained details which can help in indicating the golden signals, and troubleshoot anomalies.

Disclaimer:
{: .label .label-yellow .d-inline-block}

>**Disclaimer:** Enabling access logs may increase logs volume drastically and might fill up your log storage in a very short time. Also, the provided feature makes use of an API in alpha state which may or may not be continued in future releases.

## Prerequisites

The following resources are needed to observe access logs in you CLS setup:

- A CLS instance to ship the logs to.
- Access to the Elasticsearch API enabled on the above CLS instance.

## Configuration

### Entire Namespace

Replace `your-namespace` in the below configuration with your namespace. Use `kubectl apply` for applying the configuration.

```yaml
apiVersion: telemetry.istio.io/v1alpha1
kind: Telemetry
metadata:
  name: access-config
  namespace: "your-namespace"
spec:
  accessLogging:
    - providers:
      - name: envoy
```

### Selective Workload

Label-based selection of workloads can be configured by using [selector](https://istio.io/latest/docs/reference/config/type/workload-selector/#WorkloadSelector).
Replace `your-namespace` and `your-label` in the below configuration with your namespace and label of workload respectively.
Use `kubectl apply` for applying the configuration.

```yaml
apiVersion: telemetry.istio.io/v1alpha1
kind: Telemetry
metadata:
  name: access-config
  namespace: "your-namespace"
spec:
  selector:
    matchLabels:
      service.istio.io/canonical-name: "your-label"
```


View your logs using [Kibana](https://pages.github.tools.sap/perfx/cloud-logging-service/consumption/from-sap-cp-kyma/#view-your-logs) interface for the Cloud Logging Service.