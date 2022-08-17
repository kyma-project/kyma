---
title: External scalers
---

This tutorial shows how to use external resource scaler (like `HPA` or Keda's `ScaledObject`) with serverless Function.

Keep in mind that serverless functions implements the [scale-subresource](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#scale-subresource) which mean that you can use every `kubernetes-based` scaler instead of these described in this tutorial.

## Prerequisites

Before you start, make sure you have these tools installed:

- Kyma installed on a cluster

## Steps

Follow these steps:

<div tabs name="steps" group="create-function">
  <details>
  <summary label="hpa">
  HorizontalPodAutoscaler
  </summary>

1. Create function with the `replicas` value set to 1 to prevent internal serverless `HPA` creation:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha2
    kind: Function
    metadata:
      name: scaled-function
    spec:
      runtime: nodejs14
      replicas: 1
      source:
        inline:
          dependencies: ""
          source: |
            module.exports = {
              main: function(event, context) {
                return 'Hello World!'
              }
            }
    EOF
    ```

2. Create `HPA` using the `kubectl`:

    ```bash
    kubectl autoscale function scaled-function --cpu-percent=50 --min=5 --max=10
    ```

3. After a few seconds the `HPA` should be up to date and contain information about actual replicas:

    ```bash
    kubectl get hpa scaled-function
    ```

    Resoult should looks like this:

    ```bash
    NAME              REFERENCE                  TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
    scaled-function   Function/scaled-function   1%/50%    5         10        5          61s
    ```

  </details>
  <details>
  <summary label="keda">
  Keda
  </summary>

1. Install [Keda](https://keda.sh/docs/2.8/deploy/) if it does not present on your cluster.

2. Create function with the `replicas` value set to 1 to prevent internal serverless `HPA` creation:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha2
    kind: Function
    metadata:
      name: scaled-function
    spec:
      runtime: nodejs14
      replicas: 1
      source:
        inline:
          dependencies: ""
          source: |
            module.exports = {
              main: function(event, context) {
                return 'Hello World!'
              }
            }
    EOF
    ```

3. Create the `ScaledObject` resource:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: keda.sh/v1alpha1
    kind: ScaledObject
    metadata:
      name: scaled-function
    spec:
      scaleTargetRef:
        apiVersion:    serverless.kyma-project.io/v1alpha2
        kind:          Function
        name:          scaled-function
      minReplicaCount:  5
      maxReplicaCount:  10
      triggers:
      - type: cpu
        metricType: Utilization
        metadata:
          value: "50"
    EOF
    ```

4. After a few seconds the `ScaledObject` should be up to date and contain information about actual replicas:

    ```bash
    kubectl get scaledobject scaled-function
    ```

    Resoult should looks like this:

    ```bash
    NAME              SCALETARGETKIND                                SCALETARGETNAME   MIN   MAX   TRIGGERS   AUTHENTICATION   READY   ACTIVE   FALLBACK   AGE
    scaled-function   serverless.kyma-project.io/v1alpha2.Function   scaled-function   5     10    cpu                         True    True     Unknown    4m15s
    ```

</details>
</div>
