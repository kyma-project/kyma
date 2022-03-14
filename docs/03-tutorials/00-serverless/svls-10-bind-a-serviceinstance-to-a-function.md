---
title: Bind a Service Instance to a Function
---

>**NOTE:** This tutorial is based on the deprecated approach to Service Management in Kyma that is based on Service Catalog and Service Brokers.

This tutorial shows how you can bind a sample instance of the Redis service to a Function. After completing all steps, you will get the Function with encoded Secrets to the service. You can use them for authentication when you connect to the service to implement custom business logic of your Function.

To create the binding, you will use Service Binding and Service Binding Usage custom resources (CRs) managed by the Service Catalog.

>**NOTE:** See the document on [provisioning and binding](https://kyma-project-old.netlify.app/docs/release-1.24/components/service-catalog/#details-provisioning-and-binding-flow) to learn more about binding Service Instances to applications in Kyma.

## Prerequisites

This tutorial is based on an existing Function. To create one, follow the [Create an inline Function](./svls-01-create-inline-function.md) tutorial.

## Steps

Follow these steps:

<div tabs name="steps" group="bind-function">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Export these variables:

   ```bash
   export NAME={FUNCTION_NAME}
   export NAMESPACE={FUNCTION_NAMESPACE}
   ```

   > **NOTE:** Function takes the name from the Function CR name. The Service Instance, Service Binding, and Service Binding Usage CRs can have different names, but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

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
       - url: https://github.com/kyma-project/addons/releases/download/0.15.0/index-testing.yaml
   EOF
   ```

3. Check that the Addon CR was created successfully and that the CR phase is `Ready`:

   ```bash
   kubectl get addons $NAME -n $NAMESPACE -o=jsonpath="{.status.phase}"
   ```

4. Create a Service Instance CR. You will use the provisioned [Redis](https://redis.io/) service with its `micro` plan:

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

5. Check if the Service Instance CR was created successfully. The last condition in the CR status changes into `Ready True`:

   ```bash
   kubectl get serviceinstance $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

6. Create a Service Binding CR that points to the newly created Service Instance in the **spec.instanceRef** field:

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

7. Check if the Service Binding CR was created successfully. The last condition in the CR status changes into `Ready True`:

   ```bash
   kubectl get servicebinding $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

8. Create a Service Binding Usage CR:

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

   - The **spec.serviceBindingRef** and **spec.usedBy** fields are required. **spec.serviceBindingRef** points to the Service Binding you have just created and **spec.usedBy** points to the Function. More specifically, **spec.usedBy** refers to the name of the Function and the cluster-specific [UsageKind CR](https://kyma-project-old.netlify.app/docs/components/service-catalog/#custom-resource-usage-kind) (`kind: serverless-function`) that defines how Secrets should be injected to your Function when creating a Service Binding.

   - The **spec.parameters.envPrefix.name** field is optional. On creating a Service Binding, it adds a prefix to all environment variables injected in a Secret to the Function. In our example, **envPrefix** is `REDIS_`, so all environment variables will follow the `REDIS_{env}` naming pattern.

     > **TIP:** It is considered good practice to use **envPrefix**. In some cases, a Function must use several instances of a given Service Class. Prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another.

9. Check if the Service Binding Usage CR was created successfully. The last condition in the CR status changes into `Ready True`:

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

    > **NOTE:** If you added the **REDIS\_** prefix for environment variables in step 6, all variables will start with it. For example, the **PORT** variable will take the form of **REDIS_PORT**.

    </details>
    <details>
    <summary label="busola-ui">
    Kyma Dashboard
    </summary>

To create a binding, you must first create a sample Service Instance to which you can bind the Function. Follow the sections and steps to complete this tutorial.

### Provision a Redis service using an Addon

> **NOTE:** If you already have a Redis instance provisioned on your cluster, move directly to the **Bind the Function to the Service Instance** section.

Follow these steps:

1. Select a Namespace from the drop-down list in the top navigation panel where you want to provision the Redis service.

2. Go to **Configuration** > **Addons** and select **Create Addons Configuration**.

3. Enter the name of your Addons Configuration and `https://github.com/kyma-project/addons/releases/download/0.15.0/index-testing.yaml` in the **Urls** field.

4. Click **+**.
   
5. Select **Create** to confirm changes.

   You will see that the Addon has the status `Ready`.

### Create a Service Instance

1. Go to **Service Management** > **Catalog**. In the **Add-Ons** tab, you can see the list of all available Addons. Select **[Experimental] Redis**.

2. Select **Add** to provision the Redis Service Class and create its instance in your Namespace.

3. Change the **Name** to match the Function, select `Micro` from the **Plan** drop-down list, and set **Image pull policy** to `Always`.

   > **NOTE:** The Service Instance, Service Binding, and Service Binding Usage can have different names than the Function, but it is recommended that all related resources share a common name.

4. Select **Create** to confirm changes.

   Wait until the status of the instance changes from `PROVISIONING` to `PROVISIONED`.

### Bind the Function to the Service Instance

1. Go to **Workloads** > **Functions** and select the Function you want to bind to the Service Instance.

2. Switch to the **Configuration** tab and select **Create Service Binding** in the **Service Bindings** section.

3. Select the Redis service from the **Service Instance** drop-down list, add `REDIS_` as **Prefix for injected variables**, and make sure **Create new Secret** is selected.

4. Select **Create** to confirm changes.

The message appears on the screen confirming that the Service Binding was successfully created, and you will see it in the **Service Bindings** section in your Function, along with environment variable names.

> **NOTE:** The **Prefix for injected variables** field is optional. On creating a Service Binding, it adds a prefix to all environment variables injected in a Secret to the Function. In our example, the prefix is set to `REDIS_`, so all environment variables will follow the `REDIS_{ENVIRONMENT_VARIABLE}` naming pattern.

> **TIP:** It is considered good practice to use prefixes for environment variables. In some cases, a Function must use several instances of a given Service Class. Prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another.

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

2. Expose the Function through an [API Rule](./svls-03-expose-function.md), and access the Function's external address. You should get this result:

   ```text
   Redis port: 6379
   ```