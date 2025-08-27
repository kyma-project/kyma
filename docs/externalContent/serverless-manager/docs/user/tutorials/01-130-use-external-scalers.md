# Use External Scalers

This tutorial shows how to use an external resource scaler, for example, HorizontalPodAutoscaler (HPA) or Keda's ScaledObject, with the Serverless Function.

Keep in mind that the Serverless Functions implement the [scale subresource](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#scale-subresource), which means that you can use any Kubernetes-based scaler.

## Prerequisites

Before you start, make sure you have these tools installed:

- [Keda module added](https://kyma-project.io/docs/kyma/latest/04-operation-guides/operations/08-install-uninstall-upgrade-kyma-module/)

## Steps

Follow these steps:

<Tabs>
<Tab name="HPA">

1. Create your Function with the `replicas` value set to 1, to prevent the internal Serverless HPA creation:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha2
    kind: Function
    metadata:
      name: scaled-function
    spec:
      runtime: nodejs20
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
</Tab>
<Tab name="Keda CPU">

1. Create your Function with the **replicas** value set to `1` to prevent the internal Serverless HPA creation:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha2
    kind: Function
    metadata:
      name: scaled-function
    spec:
      runtime: nodejs20
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

2. Create the ScaledObject resource:

    ```bash
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

    > [!NOTE]
    > This tutorial uses the `cpu` trigger because of its simple configuration. If you want to use another trigger, check the official [list of supported triggers](https://keda.sh/docs/scalers/).

3. After a few seconds, ScaledObject should be up to date and contain information about the actual replicas:

    ```bash
    kubectl get scaledobject scaled-function
    ```

    You should get a result similar to this example:

    ```bash
    NAME              SCALETARGETKIND                                SCALETARGETNAME   MIN   MAX   TRIGGERS   AUTHENTICATION   READY   ACTIVE   FALLBACK   AGE
    scaled-function   serverless.kyma-project.io/v1alpha2.Function   scaled-function   5     10    cpu                         True    True     Unknown    4m15s
    ```
</Tab>
<Tab name="Keda Prometheus">

1. Create your Function with the **replicas** value set to `1` to prevent the internal Serverless HPA creation:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: serverless.kyma-project.io/v1alpha2
    kind: Function
    metadata:
      name: scaled-function
    spec:
      runtime: nodejs20
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

2. Create the ScaledObject resource based on the `istio_requests_total` metric, exposed by the Istio:

    ```bash
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
      minReplicaCount:  1  # You can go with 0 ( scaling to zero ) in case your function is fed from messaging queue that would buffer unhandled requests or if you are fine with function downtime at cold start periods
      maxReplicaCount:  5
      triggers:
      - type: prometheus
        metadata:
          serverAddress: http://prometheus-operated.kyma-system.svc.cluster.local:9090
          metricName: istio_requests_total
          query: round(sum(irate(istio_requests_total{reporter=~"source",destination_service=~"scaled-function.default.svc.cluster.local"}[2m])), 0.001)
          threshold: '6.5'
    EOF
    ```

    > [!NOTE]
    > This tutorial uses the `prometheus` trigger because of its simple configuration. If you want to use another trigger, check the official [list of supported triggers](https://keda.sh/docs/scalers/).
  
3. After a few seconds, ScaledObject should be up to date and contain information about the actual replicas:

    ```bash
    kubectl get scaledobject scaled-function
    ```

    You should get a result similar to this example:

    ```bash
    NAME              SCALETARGETKIND                                SCALETARGETNAME   MIN   MAX   TRIGGERS     AUTHENTICATION   READY   ACTIVE   FALLBACK   AGE
    scaled-function   serverless.kyma-project.io/v1alpha2.Function   scaled-function   1     5     prometheus                    True    True     Unknown      4m15s
    ```

Check out this [example](https://github.com/kyma-project/keda-manager/tree/main/examples/scale-to-zero-with-keda) to see how to use Kyma Serverless and Eventing in combination with Keda to accomplish scaling to zero.
</Tab>
</Tabs>
