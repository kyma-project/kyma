---
title: Use external scalers
---

This tutorial shows how to use an external resource scaler, for example, HorizontalPodAutoscaler (HPA) or Keda's ScaledObject, with the Serverless Function.

Keep in mind that the Serverless Functions implement the [scale subresource](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#scale-subresource), which means that you can use any Kubernetes-based scaler.

## Prerequisites

Before you start, make sure you have these tools installed:

- Kyma installed on a cluster

## Steps

Follow these steps:

<div tabs name="steps" group="create-function">
  <details>
  <summary label="hpa">
  HPA
  </summary>

1. Create your Function with the `replicas` value set to 1, to prevent the internal Serverless HPA creation:

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

2. Create your HPA using kubectl:

    ```bash
    kubectl autoscale function scaled-function --cpu-percent=50 --min=5 --max=10
    ```

3. After a few seconds your HPA should be up to date and contain information about the actual replicas:

    ```bash
    kubectl get hpa scaled-function
    ```

    You should get a result similar to this example:

    ```bash
    NAME              REFERENCE                  TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
    scaled-function   Function/scaled-function   1%/50%    5         10        5          61s
    ```

  </details>
  <details>
  <summary label="keda">
  Keda CPU
  </summary>

1. Install [Keda](https://keda.sh/docs/2.8/deploy/) if it is not present on your cluster.

2. Create your Function with the `replicas` value set to 1, to prevent the internal Serverless HPA creation:

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

3. Create the ScaledObject resource:

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

    >**NOTE:** in this tutorial we use the `cpu` trigger because of its simple configuration. If you want to use another trigger check the official [list of supported triggers](https://keda.sh/docs/2.8/scalers/).

4. After a few seconds ScaledObject should be up to date and contain information about actual replicas:

    ```bash
    kubectl get scaledobject scaled-function
    ```

    You should get a result similar to this example:

    ```bash
    NAME              SCALETARGETKIND                                SCALETARGETNAME   MIN   MAX   TRIGGERS   AUTHENTICATION   READY   ACTIVE   FALLBACK   AGE
    scaled-function   serverless.kyma-project.io/v1alpha2.Function   scaled-function   5     10    cpu                         True    True     Unknown    4m15s
    ```

  </details>
  <details>
  <summary label="keda">
  Keda CPU
  </summary>

1. Install [Keda](https://keda.sh/docs/2.8/deploy/) if it is not present on your cluster.

2. Create your Function with the `replicas` value set to 1, to prevent the internal Serverless HPA creation:

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

3. Create the ScaledObject resource:

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
      - type: prometheus
        metadata:
          serverAddress: http://prometheus-operated.kyma-system.svc.cluster.local:9090
          metricName: istio_requests_total
          query: sum(rate(istio_requests_total{destination_service_namespace="default", destination_service_name="scaled-function"}[2m]))
          threshold: '6.5'
          activationThreshold: '0'
    EOF
    ```

</details>
</div>
