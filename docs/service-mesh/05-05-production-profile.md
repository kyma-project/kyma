---
title: Service Mesh production profile
type: Configuration
---

By default, every Kyma deployment is installed with the Service Mesh provider Istio, using what is considered a development profile. In this case, this means that:
  - Horizontal Pod Autoscaler (HPA) is enabled for all components, with the default number for replicas.
  - All components have severely reduced resource quotas, which is a performance factor.

This configuration is not considered production-ready. To use the Kyma Service Mesh in a production environment, configure Istio to use the production profile.

## The production profile

The production profile introduces the following changes to the Istio Service Mesh:
   - Resource quotas for all Istio components are increased.
   - All Istio components have HPA enabled.
   - HPA for Istio Ingress Gateway has been customized:
     + Minimum number of replicas: `3`
     + Maximum number of replicas: `10`

## System requirements
As the production profile is configured with increased performance it mind, the recommend system setup is different:

| Requirement | Development setup | Production setup|
|:--- | :--- | :--- |
| vCPU | 4 | 8 |
| RAM | 16 | 16/32 |
| Example machine type (GKE) | `n1-standard-4` | `n1-standard-8` or `c2-standard-8` |
| Example machine type (AKS) | `Standard_D4_v3` | `Standard_F8s_v2` or `Standard_D8_v3` |
| Size | 3 | 3-5 |

## Use the production profile

>**CAUTION:** Due to changes in the installation options in Istio, Helm-based configuration is now deprecated in favor of the new IstioControlPlane API. Please keep in mind that Helm overrides will be no longer supported in future Istio releases. Refer to [IstioControlPlane](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) documentation for details.

You can deploy a Kyma cluster with Istio configured to use the production profile, or configure Istio in a running cluster to use the production profile. Follow these steps:

<div tabs>
  <details>
  <summary>
  Istio Control Plane API
  </summary>
Istio installation in Kyma uses the [IstioControlPlane](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) API.
This API is in the alpha version, but it's going to replace Helm-based approach in future Istio versions.
Kyma provides the default IstioControlPlane configurations for local (Minikube) and cluster installations.
You can add a custom control plane definition that overrides the default settings.
The definition you provide may be a partial one (you don't have to specify all options). In that case it will be merged with the defaults.
In order to provide a custom IstioControlPlane configuration, define a Kyma Installation override with the `kyma_istio_control_plane` key.
The value for this override must be a single string containing a valid definition of the IstioControlPlane custom resource, in the YAML format.

>**TIP:** To learn more about how to use overrides in Kyma, see the following documents:
>* [Helm overrides for Kyma installation](/root/kyma/#configuration-helm-overrides-for-kyma-installation)
>* [Top-level charts overrides](/root/kyma/#configuration-helm-overrides-for-kyma-installation-top-level-charts-overrides).

See the following example that customizes settings for the `policy` and `pilot` components of Istio:

    ```bash
    cat <<EOF | kubectl apply -f -
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: istio-control-plane-overrides
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
          policy:
            components:
              policy:
                enabled: true
                k8s:
                  replicaCount: 1
                  resources:
                    limits:
                      cpu: 543m
                      memory: 2048Mi
                    requests:
                      cpu: 321m
                      memory: 512Mi
                  strategy:
                    rollingUpdate:
                      maxSurge: 1
                      maxUnavailable: 0
            enabled: true
          trafficManagement:
            components:
              pilot:
                enabled: true
                k8s:
                  affinity:
                    podAntiAffinity:
                      preferredDuringSchedulingIgnoredDuringExecution: []
                      requiredDuringSchedulingIgnoredDuringExecution: []
                  env:
                    - name: GODEBUG
                      value: gctrace=1
                    - name: PILOT_HTTP10
                      value: "1"
                    - name: PILOT_PUSH_THROTTLE
                      value: "100"
                  nodeSelector: {}
                  resources:
                    limits:
                      cpu: 567m
                      memory: 1024Mi
                    requests:
                      cpu: 234m
                      memory: 512Mi
                  strategy:
                    rollingUpdate:
                      maxSurge: 1
                      maxUnavailable: 0
                  tolerations: []
            enabled: true
    EOF
    ```

Refer to the [IstioControlPlane API](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) documentation for details about available options.
  </details>
  <details>
  <summary>
  Install Kyma with production-ready Istio
  </summary>

  1. Create an appropriate Kubernetes cluster for Kyma in your host environment.
  2. Apply an override that forces the Istio Service Mesh to use the production profile. Run:
    ```bash
    cat <<EOF | kubectl apply -f -
    ---
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
      global.proxy.resources.requests.cpu: "150m"
      global.proxy.resources.requests.memory: "128Mi"
      global.proxy.resources.limits.cpu: "500m"
      global.proxy.resources.limits.memory: "1024Mi"

      gateways.istio-ingressgateway.autoscaleMin: "3"
      gateways.istio-ingressgateway.autoscaleMax: "10"
    EOF
    ```
  3. Install Kyma on the cluster.

  </details>
  <details>
  <summary>
  Enable production profile in a running cluster
  </summary>

  1. Apply an override that forces the Istio Service Mesh to use the production profile. Run:
    ```bash
    cat <<EOF | kubectl apply -f -
    ---
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
      global.proxy.resources.requests.cpu: "150m"
      global.proxy.resources.requests.memory: "128Mi"
      global.proxy.resources.limits.cpu: "500m"
      global.proxy.resources.limits.memory: "1024Mi"

      gateways.istio-ingressgateway.autoscaleMin: "3"
      gateways.istio-ingressgateway.autoscaleMax: "10"
    EOF
    ```
  2. Run the [cluster update procedure](/root/kyma/#installation-update-kyma).

  </details>
</div>
