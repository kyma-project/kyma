---
title: Bind a Service Instance to a lambda
type: Tutorials
---

This tutorial shows how you can bind a sample instance of the Redis service to a lambda. After completing all steps, you will get the lambda with encoded Secrets to the service. You can use them for authentication when you connect to the service to implement custom business logic of your lambda.

To create the binding, you will use ServiceBinding and ServiceBindingUsage custom resources (CRs) managed by the Service Catalog.

>**NOTE:** To learn more about binding Service Instances to applications in Kyma, read [this](/components/service-catalog/#details-provisioning-and-binding) document.

## Prerequisites

This tutorial is based on an existing lambda. To create one, follow the [Create a lambda](#tutorials-create-a-lambda) tutorial.

## Steps

Follows these steps:

1. Export these variables:

    ```bash
    export NAME={LAMBDA_NAME}
    export NAMESPACE=serverless
    ```

    > **NOTE:** Lambda takes the name from the Function CR name. The ServiceInstance, ServiceBinding, and ServiceBindingUsage CRs can have different names, but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

2. Provision an Addon CR with the Redis service:

    ```yaml
    cat <<EOF | kubectl apply -f  -
    apiVersion: addons.kyma-project.io/v1alpha1
    kind: AddonsConfiguration
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      reprocessRequest: 0
      repositories:
      - url: https://github.com/kyma-project/addons/releases/download/0.11.0/index-testing.yaml
    EOF
    ```

3. Check if the Addon CR was created successfully. The CR phase should state `Ready`:

    ```bash
    kubectl get addons $NAME -n $NAMESPACE -o=jsonpath="{.status.phase}"
    ```

4. Create a ServiceInstance CR. You will use the provisioned [Redis](https://redis.io/) service with its `micro` plan:

    ```yaml
    cat <<EOF | kubectl apply -f -
    apiVersion: servicecatalog.k8s.io/v1beta1
    kind: ServiceInstance
    metadata:
      name: $NAME
      namespace: $NAMESPACE
    spec:
      serviceClassExternalName: redis
      servicePlanExternalName: micro
      parameters:
        imagePullPolicy: Always
    EOF    
    ```

5. Check if the ServiceInstance CR was created successfully. The last condition in the CR status should state `Ready True`:

    ```bash
    kubectl get serviceinstance $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
    ```

6. Create a ServiceBinding CR that points to the newly created Service Instance in the **spec.instanceRef** field:

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

    > **NOTE:** If you use an existing Service Instance, change **spec.instanceRef.name** to the name of your Service Instance.

7. Check if the ServiceBinding CR was created successfully. The last condition in the CR status should state `Ready True`:

    ```bash
    kubectl get servicebinding $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
    ```

8. Create a ServiceBindingUsage CR:

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

    - The **spec.serviceBindingRef** and **spec.usedBy** fields are required. **spec.serviceBindingRef** points to the Service Binding you have just created and **spec.usedBy** points to the lambda. More specifically, **spec.usedBy** refers to the name of the related KService CR (`name: $NAME`) and the cluster-specific [UsageKind CR](https://kyma-project.io/docs/components/service-catalog/#custom-resource-usage-kind) (`kind: knative-service`) that defines how Secrets should be injected to your lambda through the Service Binding.

    > **NOTE:** Read [this](/components/service-catalog/#details-provisioning-and-binding) document to learn more about binding in Kyma.

    - The **spec.parameters.envPrefix.name** field is optional. It adds a prefix to all environment variables injected by a given Secret from the Service Binding to the lambda. In our example, **envPrefix** is `REDIS_`, so all environmental variables will follow the `REDIS_{env}` naming pattern.

    > **TIP:** It is considered good practice to use **envPrefix**. In some cases, lambda must use several instances of a given Service Class, so prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another one.

9. Check if the ServiceBindingUsage CR was created successfully. The last condition in the CR status should state `Ready True`:

    ```bash
    kubectl get servicebindingusage $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
    ```

10. Retrieve and decode Secret details from the Service Binding:

    ```bash
    kubectl get secret $NAME -n $NAMESPACE -o go-template='{{range $k,$v := .data}}{{printf "%s: " $k}}{{if not $v}}{{$v}}{{else}}{{$v | base64decode}}{{end}}{{"\n"}}{{end}}'
    ```

    You should get a result similar to the following details:

    ```bash
    HOST: hb-redis-micro-0e965585-9699-443f-b987-38bc6af0e416-redis.serverless.svc.cluster.local
    PORT: 6379
    REDIS_PASSWORD: 1tvDcINZvp
    ```

    > **NOTE:** If you added the **REDIS_** prefix for environmental variables in step 6, all variables will start with it. For example, the **PORT** variables will take the form of **REDIS_PORT**.
