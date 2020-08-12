---
title: Bind a Redis ServiceInstance to a Function
type: Getting Started
---

This tutorial shows how you can bind a sample instance of the Redis service to a Function. After completing all steps, you will get the Function with encoded Secrets to the service. You can use them for authentication when you connect to the service to implement custom business logic of your Function.

## Prerequisites

Follow these tutorials first to:
1. [Create a Function](/#tutorials-create-a-function)
2. [Provision the Redis Addon](/#tutorials-register-an-addon)
3. [Create a ServiceInstance](/#tutorials-create-a-service-instance)

## Steps

Follows these steps:

<div tabs name="steps" group="bind-function">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Export these variables:

   ```bash
   export NAME={FUNCTION_NAME}
   export NAMESPACE={FUNCTION_NAMESPACE}
   ```

2. Create a ServiceBinding CR that points to the newly created Service Instance in the **spec.instanceRef** field:

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

3. Check if the ServiceBinding CR was created successfully. The last condition in the CR status should state `Ready True`:

   ```bash
   kubectl get servicebinding $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

4. Create a ServiceBindingUsage CR:

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
       kind: serverless-function
       name: $NAME
     parameters:
       envPrefix:
         name: "REDIS_"
   EOF
   ```

   - The **spec.serviceBindingRef** and **spec.usedBy** fields are required. **spec.serviceBindingRef** points to the Service Binding you have just created and **spec.usedBy** points to the Function. More specifically, **spec.usedBy** refers to the name of the Function and the cluster-specific [UsageKind CR](https://kyma-project.io/docs/components/service-catalog/#custom-resource-usage-kind) (`kind: serverless-function`) that defines how Secrets should be injected to your Function when creating a Service Binding.

   - The **spec.parameters.envPrefix.name** field is optional. It adds a prefix to all environment variables injected in a Secret to the Function when creating a Service Binding. In our example, **envPrefix** is `REDIS_`, so all environmental variables will follow the `REDIS_{env}` naming pattern.

     > **TIP:** It is considered good practice to use **envPrefix**. In some cases, a Function must use several instances of a given ServiceClass. Prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another one.

5. Check if the ServiceBindingUsage CR was created successfully. The last condition in the CR status should state `Ready True`:

   ```bash
   kubectl get servicebindingusage $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

6. Retrieve and decode Secret details from the Service Binding:

    ```bash
    kubectl get secret $NAME -n $NAMESPACE -o go-template='{{range $k,$v := .data}}{{printf "%s: " $k}}{{if not $v}}{{$v}}{{else}}{{$v | base64decode}}{{end}}{{"\n"}}{{end}}'
    ```

    You should get a result similar to the following details:

    ```bash
    HOST: hb-redis-micro-0e965585-9699-443f-b987-38bc6af0e416-redis.serverless.svc.cluster.local
    PORT: 6379
    REDIS_PASSWORD: 1tvDcINZvp
    ```

    > **NOTE:** If you added the **REDIS\_** prefix for environmental variables in step 6, all variables will start with it. For example, the **PORT** variable will take the form of **REDIS_PORT**.

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. Go to the **Functions** view in the left navigation panel and select the Function you want to bind to the Service Instance.

2. Switch to the **Configuration** tab and select **Create Service Binding** in the **Service Bindings** section.

3. Select the Redis service from the **Service Instance** drop-down list, add `REDIS_` as **Prefix for injected variables**, and make sure **Create new Secret** is selected.

4. Select **Create** to confirm changes.

The message appears on the screen confirming that the Service Binding was successfully created, and you will see it in the **Service Bindings** section in your Function, along with environment variable names.

> **NOTE:** The **Prefix for injected variables** field is optional. It adds a prefix to all environment variables injected in a Secret to the Function when creating a Service Binding. In our example, the prefix is set to `REDIS_`, so all environmental variables will follow the `REDIS_{ENVIRONMENT_VARIABLE}` naming pattern.

> **TIP:** It is considered good practice to use prefixes for environment variables. In some cases, a Function must use several instances of a given ServiceClass. Prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another one.

    </details>

</div>

## Test the Function

To test if the Secret has been properly connected to the Function:

1. Change the Function's code to:​

   ```js
   module.exports = {
     main: function (event, context) {
       return "Redis port: " + process.env.REDIS_PORT;
     },
   };
   ```

2. Expose the Function through an [API Rule](#tutorials-expose-a-function-with-an-api-rule), and access the Function's external address. You should get this result:

   ```text
   Redis port: 6379
   ```
