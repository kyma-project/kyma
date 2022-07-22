---
title: Jaeger doesn't show the traces you want to see
---

## Condition

Jaeger shows fewer traces than you would like to see.

## Cause

By default, only 1% of the requests are sent to Jaeger for trace recording.

## Remedy

To see more traces in Jaeger, increase the percentage of requests by changing the default settings.
If you just want to see traces for one particular request, you can manually force sampling.

### Change the default setting

To override the default percentage, you deploy a YAML file. You can do this either during initial installation or to adjust an existing Kyma installation.

1. To set the value for the **trace sampling** attribute, create a values YAML file.
   The following example sets the value to `60`, which means 60% of the requests are sent to Jaeger.

   ```yaml
    istio:
      meshConfig:
        defaultConfig:
          tracing:
            sampling: 60
   ```

   > **CAUTION:** Sending 100% of the requests to Jaeger might destabilize Istio.

2. Deploy the values YAML file with the following command:

   ```bash
   kyma deploy --values-file {VALUES_FILE_PATH}
   ```

### Force sampling for a particular request

If you want to force sampling for a particular request, set the `x-b3-sampled: 1` http header manually in the application code.
