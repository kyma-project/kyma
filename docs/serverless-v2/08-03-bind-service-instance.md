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

<div tabs name="steps" group="bind-lambda">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Export these variables:

    ```bash
    export NAME={LAMBDA_NAME}
    export NAMESPACE={LAMBDA_NAMESPACE}
    ```

    > **NOTE:** Lambda takes the name from the Function CR name. The ServiceInstance, ServiceBinding, and ServiceBindingUsage CRs can have different names, but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

    > **NOTE:** If you already have a Redis instance provisioned on your cluster, move directly to point 6 to create a Service Binding.

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

    - The **spec.parameters.envPrefix.name** field is optional. It adds a prefix to all environment variables injected by a given Secret from the Service Binding to the lambda. In our example, **envPrefix** is `REDIS_`, so all environmental variables will follow the `REDIS_{env}` naming pattern.

        > **TIP:** It is considered good practice to use **envPrefix**. In some cases, a lambda must use several instances of a given Service Class. Prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another one.

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

    > **NOTE:** If you added the **REDIS_** prefix for environmental variables in step 6, all variables will start with it. For example, the **PORT** variable will take the form of **REDIS_PORT**.

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

> **NOTE:** Serverless v2 is an experimental feature, and it is not enabled by default in the Console UI. To use its **Functions [preview]** view, enable **Experimental functionalities** in the **General Settings** view before you follow the steps.

To create a binding, you must first create a sample service instance to which you can bind the lambda. Follow the sections and steps to complete this tutorial.

### Provision a Redis service using an Addon

> **NOTE:** If you already have a Redis instance provisioned on your cluster, move directly to the **Bind the lambda with the service Instance** section.

Follow these steps:

1. Select a Namespace from the drop-down list in the top navigation panel where you want to provision the Redis service.
2. Go to the **Addons** view in the left navigation panel and select **Add New Configuration**.
3. Enter `https://github.com/kyma-project/addons/releases/download/0.11.0/index-testing.yaml` in the **Urls** field. The Addon name is automatically generated.
4. Select **Add** to confirm changes.

    You will see that the Addon has the `Ready` status.

### Create a Service Instance

1. Go to the Catalog view where you can see the list of all available Addons and select **Redis**.
2. Select **Add** to provision the Redis ServiceClass and create its instance in your Namespace.
3. Change the **Name** to match the lambda, select `micro` from the **Plan** drop-down list, and set **Image pull policy** to `Always`.

    > **NOTE:** The Service Instance, Service Binding, and Service Binding Usage can have different names than lambda, but it is recommended that all related resources share a common name.

4. Select **Create** to confirm changes.

    Wait until the status of the instance changes from `PROVISIONING` to `RUNNING`.

### Bind the lambda with the service Instance

1. Go to the **Functions [preview]** view at the bottom of the left navigation panel and select the lambda you want to bind to the Service Instance.
2. Select **Select Service Bindings** in the **Service Bindings** section.
3. Select the Redis service from the **Service Instance** drop-down list, add `REDIS_` as **Prefix for injected variables**, and make sure **Create new Secret** is checked.
4. Select **Create** to confirm changes.

  You will see the `Service Binding creating...` message and the binding available under the **Service Bindings** section in your lambda, along with **Environment Variable Names**.

The **Prefix for injected variables** field is optional. It adds a prefix to all environment variables injected by a given Secret from the Service Binding to the lambda. In our example, the prefix is set to `REDIS_`, so all environmental variables will follow the `REDIS_{env}` naming pattern.

> **TIP:** It is considered good practice to use prefixes for environment variables. In some cases, a lambda must use several instances of a given Service Class. Prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another one.

    </details>
</div>

## Test the lambda

To test if the Secret has been properly connected to the lambda:

1. Change the lambda's code to:​

    ```js
    module.exports = {
      main: function (event, context) {
        return "Redis port: " + process.env.REDIS_PORT;
      }
    }
    ```

2. Expose the lambda through an [API Rule](/components/serverless-v2/#tutorials-expose-the-lambda-with-an-api-rule), and access the lambda's external address. You should get this result:

    ```text
    Redis port: 6379
    ```
