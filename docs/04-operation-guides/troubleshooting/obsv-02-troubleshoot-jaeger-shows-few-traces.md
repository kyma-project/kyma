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

   > **CAUTION:** Be careful if you consider sending 100% of the requests to Jaeger, as this might destabilize Istio.

   ```yaml
   istio:
     kyma_istio_operator: |-
     apiVersion: install.istio.io/v1alpha1
     kind: IstioOperator
     metadata:
       namespace: istio-system
     spec:
       meshConfig:
         defaultConfig:
           tracing:
             sampling: 60
   ```

2. Deploy the values YAML file with the following command:

   ```bash
   kyma deploy --values-file {VALUES_FILE_PATH}
   ```

3. If you add the override in the Runtime, run the following command to trigger the update:

    > ```bash
    > kubectl -n default label installation/kyma-installation action=install
    > ```

   If you have already installed Kyma and do not want to trigger any updates, edit the `istiod` deployment to set the desired value for **PILOT_TRACE_SAMPLING**. For detailed instructions, see the [Istio documentation](https://istio.io/latest/docs/tasks/observability/distributed-tracing/configurability/#customizing-trace-sampling).

   >**NOTE:** Only if the meshConfig override is not defined, the change to PILOT_TRACE_SAMPLING takes effect.

### Force sampling for a particular request

You can also manually set the `x-b3-sampled: 1` header to force sampling for a particular request.
