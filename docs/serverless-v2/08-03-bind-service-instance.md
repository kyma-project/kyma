---
title: Bind a Service Instance to lambda
type: Tutorials
---

This tutorial shows how you can bind a Service Instance to lambda in Kyma. To bind it, use an ServiceBinding and ServiceBindingUsage custom resources (CRs) managed by the in-house Service Catalog domain.

When you complete this tutorial, you get a lambda that:

- Has a bound secrets that can be used to connect to the Service Instance.

>**NOTE:** To learn more about binding a Service Instances to applications in Kyma, see [this](/components/service-catalog/#details-provisioning-and-binding) document.

## Prerequisites

This tutorial is based on an existing lambda. To create lambda, follow the [Create a lambda](#tutorials-create-a-lambda) tutorial.

Additionally one Service Instance is required. In this tutorial will be provisioned and used [Redis](https://redis.io/) with micro plan.

## Steps

Follows these steps:

1. Export these variables:

    ```bash
    export DOMAIN={DOMAIN_NAME}
    export NAME={LAMBDA_NAME}
    export NAMESPACE=serverless
    ```

    > **NOTE:** Lambda takes the name from the Function CR name. The ServiceInstance, ServiceBinding and ServiceBindingUsage CRs can have a different names but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable

    > **NOTE:** If you have a Service Instance created earlier you can go to 4 point.

2. Create a Service Instance CR (is this tutorial we use Redis with micro plan).

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: servicecatalog.k8s.io/v1beta1
    kind: ServiceInstance
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      clusterServiceClassExternalName: redis
      clusterServicePlanExternalName: micro
      parameters:
        imagePullPolicy: Always
    EOF    
    ```

3. Check if the Service Instance was created successfully by checking that last of conditions is `Ready True`:

    ```bash
    kubectl get serviceinstance $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
    ```

4. Create a Service Binding CR that points in `spec.instanceRef` field to created in previous steps Service Instance.

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: servicecatalog.k8s.io/v1beta1
    kind: ServiceBinding
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      instanceRef:
        name: $NAME
    EOF    
    ```

    > **NOTE:** If you have a Service Instance created earlier, remember to change `spec.instanceRef.name` to your Service Instance name.

5. Check if the Service Binding was created successfully by checking that last of conditions is `Ready True`:

    ```bash
    kubectl get servicebinding $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
    ```

6. Create a Service Binding Usage CR.

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: servicecatalog.kyma-project.io/v1alpha1
    kind: ServiceBindingUsage
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      serviceBindingRef:
        name: $NAME
      usedBy:
        kind: knative-service
        name: $NAME
      parameters:
        envPrefix:
          name: "REDIS_"
    EOF    
    ```

    `spec.serviceBindingRef` and `spec.usedBy` fields are required. `spec.serviceBindingRef` points to created in previous steps Service Binding and `spec.usedBy` points to lambda.

    `spec.parameters.envPrefix.name` field is optional, which adds a prefix to all environment variables injected from a given Secret from Service Binding to lambda. In our example `envPrefix` is `REDIS_`, so environmental variables will have a form `REDIS_{env}`.

    > **NOTE:** `spec.usedBy` field points to Knative Service with `$NAME` name (by `knative-service` kind), not to Function CR, because secret is bound at the Pod level. Read more about binding in Kyma [here](/components/service-catalog/#details-provisioning-and-binding).

7. Check if the Service Binding Usage was created successfully by checking that last of conditions is `Ready True`:

    ```bash
    kubectl get servicebindingusage $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
    ```

8. Now you can see what environment variables you can use in lambda. Run below command to retrieve and decode secret data from created Service Binding:

    ```bash
    kubectl get secret $NAME -n $NAMESPACE -o go-template='{{range $k,$v := .data}}{{printf "%s: " $k}}{{if not $v}}{{$v}}{{else}}{{$v | base64decode}}{{end}}{{"\n"}}{{end}}'
    ```

    You should see something like:

    ```bash
    HOST: hb-redis-micro-0e965585-9699-443f-b987-38bc6af0e416-redis.serverless.svc.cluster.local
    PORT: 6379
    REDIS_PASSWORD: 1tvDcINZvp
    ```

    > **NOTE:** Remember about prefix for environmental variables defined in previous steps. For example: `PORT` env will have a form `REDIS_PORT` etc.

    Changing lambda's source to something like below, function should return in response `hb-redis-micro-0e965585-9699-443f-b987-38bc6af0e416-redis.serverless.svc.cluster.local`.

    ```js
    module.exports = {
      main: function(event, context) {
        return process.env.REDIS_HOST;
      }
    }
    ```
    