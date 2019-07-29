---
title: Basic Troubleshooting
type: Troubleshooting
---

## Jaeger shows only a few traces

The current Istio Pilot settings define the trace sampling rate at `1.0`, where `100` is the maximum value. This means that only 1 out of 100 requests is sent to Jaeger for trace recording. To change this system behavior, run:

```bash
kubectl -n istio-system edit deploy istio-pilot
```
Set the **traceSampling** parameter to a desired value, such as `60`. 

>**NOTE**: Using a very high value may affect Istio's performance and stability. 