---
title: Trace backend doesn't show the traces you want to see
---

## Condition

Trace backend shows fewer traces than you would like to see.

## Cause

By [default](https://kyma-project.io/docs/kyma/latest/01-overview/telemetry/telemetry-03-traces#istio), only 1% of the requests are sent to the trace backend for trace recording.

## Remedy

To see more traces in the trace backend, increase the percentage of requests by changing the default settings.
If you just want to see traces for one particular request, you can manually force sampling.

### Change the default setting

To override the default percentage, you deploy a YAML file to an existing Kyma installation.

1. To set the value for the **randomSamplingPercentage** attribute, create a values YAML file.
   The following example sets the value to `60`, which means 60% of the requests are sent to tracing backend.

   ```yaml
    apiVersion: telemetry.istio.io/v1alpha1
    kind: Telemetry
    metadata:
      name: kyma-traces
      namespace: istio-system
    spec:
      tracing:
      - providers:
        - name: "kyma-traces"
        randomSamplingPercentage: 60
   ```
