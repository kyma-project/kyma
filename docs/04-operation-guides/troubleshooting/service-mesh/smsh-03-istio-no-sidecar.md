---
title: Issues with Istio sidecar injection
---

## Symptom

A Pod doesn't have a sidecar but you did not disable sidecar injection on purpose.

## Cause

By default, Kyma has sidecar injection disabled - there is no automatic sidecar injection into any Pod in a cluster. For more information, read the document about [enabling Istio sidecar proxy injection](../../operations/smsh-01-istio-enable-sidecar-injection.md).

## Remedy

Follow these steps to troubleshoot:

1. Check if sidecar injection is enabled in the Namespace of the Pod. Run this command to check the `istio-injection` label:

    ```bash
    kubectl get namespaces {NAMESPACE} -o jsonpath='{ .metadata.labels.istio-injection }'
    ```

   If the command does not return `enabled`, the sidecar injection is disabled in this Namespace. To add a sidecar to the Pod, move the Pod's deployment to a Namespace that has sidecar injection enabled, or add the label to the Namespace and restart the Pod.

   >**WARNING:** Adding the `istio-injection=enabled` label from the Namespace results in injecting sidecars to all Pods inside of the Namespace.

2. Check if sidecar injection is enabled in the Pod's Deployment:

    ```bash
    kubectl get deployments {DEPLOYMENT_NAME} -n {NAMESPACE} -o jsonpath='{ .spec.template.metadata.labels }'
    ```

   Sidecar injection is disabled if the output does not contain the `sidecar.istio.io/inject:true` line. Change or add the label and restart the Pod to enable sidecar injection for the Deployment.

For more information, read [Istio documentation](https://istio.io/docs/ops/common-problems/injection/).
