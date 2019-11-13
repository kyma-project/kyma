---
title: Basic Troubleshooting
type: Troubleshooting
---

## Jaeger shows only a few traces

1. The current Istio Pilot settings define the trace sampling rate at `1.0`, where `100` is the maximum value. This means that only 1 out of 100 requests is sent to Jaeger for trace recording. To change this system behavior, run:

```bash
kubectl -n istio-system edit deploy istio-pilot
```
Set the **traceSampling** parameter to a desired value, such as `60`.

>**NOTE**: Using a very high value may affect Istio's performance and stability.

2. The current Knative trace sampling rate is also at `0.1` where `1` is the maximum value. To have complete trace recordings, run:

```bash
kubectl edit cm -n knative-eventing config-tracing
```
Set the **sample-rate** parameter to a desired value, such as `1`.

>**NOTE**: Using a very high value may affect the memory usage of Jaeger's deployment so increasing the memory limits is needed.
