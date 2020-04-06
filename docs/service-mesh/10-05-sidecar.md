---
title: Issues with Istio sidecar injection
type: Troubleshooting
---


Kyma has sidecar injection enabled by default - it injects sidecars to every Deployment on a cluster, without the need of labeling a Deployment or Namespace. For more information, read [this document](#details-sidecar-proxy-injection).
If your Pod doesn't have a sidecar, although it should, follow this steps:

1. Check if sidecar injection is disabled in the Namespace of the Pod:

    ```bash
    kubectl get namespaces {NAMESPACE} -o jsonpath='{ .metadata.labels.istio-injection }'
    ```

    Sidecar injection is disabled in this namespace if the output is `disabled`. Move Deployment to another Namespace or delete the label and restart the Pod.
    
    >**WARNING:** Removing the label from Namespace will result with injecting sidecars to all Pods inside of the Namespace.
  
2. Check if sidecar injection is disabled in the Deployment:

    ```bash
    kubectl get deployments {DEPLOYMENT_NAME} -n {NAMESPACE} -o jsonpath='{ .spec.template.metadata.annotations }'
    ```
   
   Sidecar injection is disabled if the output contains a line `sidecar.istio.io/inject:false`. Delete the label and restart the Pod.
   
3. Make sure Istio sidecar injector is running: 
    
    ```bash
    kubectl describe pod -n istio-system -l app=sidecarInjectorWebhook
    ```

4. Make sure Istio sidecar injector can communicate with the Kubernetes API server. Search logs for any issues regarding connectivity:

    ```bash
    kubectl logs -n istio-system -l app=sidecarInjectorWebhook
    ```

For more information, read [this document](#details-sidecar-proxy-injection) or follow the [Istio documentation](https://istio.io/docs/ops/common-problems/injection/).
