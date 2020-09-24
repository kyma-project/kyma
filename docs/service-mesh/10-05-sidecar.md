---
title: Issues with Istio sidecar injection
type: Troubleshooting
---


Kyma has sidecar injection enabled by default - a sidecar is injected to every Deployment in a cluster, without the need adding any labels. For more information, read [this document](#details-sidecar-proxy-injection).
If a Pod doesn't have a sidecar and you did not disable sidecar injection on purpose, follow these steps to troubleshoot:

1. Check if sidecar injection is disabled in the Namespace of the Pod. Run this command to check the `istio-injection` label:

    ```bash
    kubectl get namespaces {NAMESPACE} -o jsonpath='{ .metadata.labels.istio-injection }'
    ```

    If the command returns `disabled` the sidecar injection is disabled in this Namespace. To add a sidecar to the Pod, move the Pod's deployment to a Namespace that has sidecar injection enabled, or remove the label from the Namespace and restart the Pod.
    
    >**WARNING:** Removing the `istio-injection=disabled` label from Namespace results in injecting sidecars to all Pods inside of the Namespace.
  
2. Check if sidecar injection is disabled in the Pod's Deployment:

    ```bash
    kubectl get deployments {DEPLOYMENT_NAME} -n {NAMESPACE} -o jsonpath='{ .spec.template.metadata.annotations }'
    ```
   
   Sidecar injection is disabled if the output contains the `sidecar.istio.io/inject:false` line. Delete the label and restart the Pod to enable sidecar injection for the Deployment.
   

For more information, read [this document](#details-sidecar-proxy-injection) or follow the [Istio documentation](https://istio.io/docs/ops/common-problems/injection/).
