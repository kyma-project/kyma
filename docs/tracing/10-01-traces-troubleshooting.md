---
title: Jaeger shows only a few traces
type: Troubleshooting
---

By default, only 1% of the requests are sent to Jaeger for trace recording. To change this system behavior, you can override the existing settings or change the value in the Runtime.

> **NOTE:** You can also manually set the `x-b3-sampled: 1` header to force sampling for a particular request.

## Create an override

Follow these steps to [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation) the existing configuration

1. Add and apply a ConfigMap in the `kyma-installer` Namespace in which you set the value for the **trace sampling** attribute to `60`.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: istio-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: istio
    kyma-project.io/installation: ""
data:
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
EOF
```

2. Proceed with the installation. Once the installation starts, the Kyma Operator will generate the override based on the ConfigMap and set the value of trace sampling to `60`.

    >**NOTE:** If you add the override in the Runtime, run the following command to trigger the update:
    > ```bash
    > kubectl -n default label installation/kyma-installation action=install
    > ```

## Define the value in the Runtime

If you have already installed Kyma and do not want to trigger any updates, edit the `istiod` deployment to set the desired value for **PILOT_TRACE_SAMPLING**. For detailed instructions, see the [Istio documentation](https://istio.io/latest/docs/tasks/observability/distributed-tracing/configurability/#customizing-trace-sampling).

> Note: the change to PILOT_TRACE_SAMPLING would only take effect if the meshConfig override is not defined.
