---
title: Create a ServiceInstance for the Redis service
type: Getting Started
---

To create a binding between a Function and a provisioned Service, you must first create a sample Service Instance to which you can bind the Function. Follow the sections and steps to complete this tutorial.

To create the binding, you will use ServiceBinding and ServiceBindingUsage custom resources (CRs) managed by the [Service Catalog](https://svc-cat.io/docs/walkthrough/).

>**NOTE:** See the document on [provisioning and binding](/components/service-catalog/#details-provisioning-and-binding) to learn more about binding Service Instances to applications in Kyma.

## Prerequisites

To follow this tutorial you must first [provision the Redis service](/#tutorials-register-an-addon) with its `micro` plan.

## Steps

Follows these steps:

<div tabs name="steps" group="create-instance">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Export these variables:

   ```bash
   export NAME={FUNCTION_NAME}
   export NAMESPACE={FUNCTION_NAMESPACE}
   ```

   > **NOTE:** Function takes the name from the Function CR name. The ServiceInstance, ServiceBinding, and ServiceBindingUsage CRs can have different names, but for the purpose of this tutorial, all related resources share a common name defined under the **NAME** variable.

2. Create a ServiceInstance CR. You will use the provisioned [Redis](https://redis.io/) service with its `micro` plan:

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

3. Check if the ServiceInstance CR was created successfully. The last condition in the CR status should state `Ready True`:

   ```bash
   kubectl get serviceinstance $NAME -n $NAMESPACE -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

### Create a Service Instance

1. Go to the **Catalog** view where you can see the list of all available Addons and select **[Experimental] Redis**.

2. Select **Add** to provision the Redis ServiceClass and create its instance in your Namespace.

3. Change the **Name** to match the Function, select `micro` from the **Plan** drop-down list, and set **Image pull policy** to `Always`.

   > **NOTE:** The Service Instance, Service Binding, and Service Binding Usage can have different names than the Function, but it is recommended that all related resources share a common name.

4. Select **Create** to confirm changes.

   Wait until the status of the instance changes from `PROVISIONING` to `RUNNING`.


    </details>
</div>
