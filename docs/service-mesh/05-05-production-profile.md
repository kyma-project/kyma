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

You can deploy a Kyma cluster with Istio configured to use the production profile, or configure Istio in a running cluster to use the production profile. Follow these steps:

<div tabs>
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
      global.proxy.resources.requests.cpu: "300m"
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
      global.proxy.resources.requests.cpu: "300m"
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
