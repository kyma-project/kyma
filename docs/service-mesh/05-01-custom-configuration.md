---
title: Istio custom configuration
type: Configuration
---
The Istio installation in Kyma uses the [IstioOperator](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) API.
Kyma provides the default IstioOperator configurations for local (Minikube) and cluster installations, but you can add a custom IstioOperator definition that overrides the default settings.

The definition you provide may be a partial one with not all the options specified. In that case, it will be merged with the defaults.

To provide a custom IstioOperator configuration, define a Kyma Installation override with the **kyma_istio_operator** key.
The value for this override must be a single string containing a valid definition of the IstioOperator custom resource, in the YAML format.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides).

See the following example that customizes settings for the `policy` and `pilot` components of Istio:

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: istio-operator-overrides
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
          name: example-istiooperator
        spec:
          components:
            policy:
              k8s:
                hpaSpec:
                  minReplicas: 2
            pilot:
              k8s:
                resources:
                  requests:
                    memory: 3072Mi
                env:
                - name: PILOT_TRACE_SAMPLING
                  value: "20"
    ```

While installing with Kyma CLI, don't forget to provide this file's path via `-o` flag.

Refer to the [IstioOperator API](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) documentation for details about available options.
