---
title: Basic Troubleshooting
type: Troubleshooting
---

## Jaeger shows only a few traces

By default, the **PILOT_TRACE_SAMPLING** value in the [IstioControlPlane](https://istio.io/docs/reference/config/istio.operator.v1alpha12.pb/) is set to `1`, where `100` is the maximum value. This means that only 1 out of 100 requests is sent to Jaeger for trace recording.
To change this system behavior, [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation) the existing configuration with a custom control plane definition.

1. Add and apply a ConfigMap in the `kyma-installer` Namespace in which you set the value for the **PILOT_TRACE_SAMPLING** attribute to `60`.

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
  kyma_istio_control_plane: |-
    apiVersion: install.istio.io/v1alpha2
    kind: IstioControlPlane
    spec:
      trafficManagement:
        components:
          pilot:
            enabled: true
            k8s:
              env:
              - name: PILOT_TRACE_SAMPLING
                value: "60"
EOF
```

2. Proceed with the installation. Once the installation starts, the Kyma Operator will generate the override based on the ConfigMap and set the value of **PILOT_TRACE_SAMPLING** to `60`.

    >**NOTE:** If you use the override at runtime, run the following command to trigger the installation:
    > ```bash
    > kubectl -n default label installation/kyma-installation action=install
    > ```
