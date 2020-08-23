---
title: Create a ServiceInstance for the Redis service
type: Getting Started
---

To create a binding between the microservice and a provisioned Redis service, you must first create an instance of this service to bind the microservice to it.

## Prerequisites

To follow this tutorial you must first [provision the Redis service](/#tutorials-register-an-addon) with its `micro` plan.

## Steps

Follows these steps:

<div tabs name="steps" group="create-redis-instance">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Create a ServiceInstance CR. You will provision the [Redis](https://redis.io/) service with its `micro` plan:

   ```yaml
   cat <<EOF | kubectl apply -f -
   apiVersion: servicecatalog.k8s.io/v1beta1
   kind: ServiceInstance
   metadata:
     name: redis-service
     namespace: orders-service
   spec:
     serviceClassExternalName: redis
     servicePlanExternalName: micro
     parameters:
       imagePullPolicy: Always
   EOF
   ```

2. Check if the ServiceInstance CR was created successfully. The last condition in the CR status should state `Ready True`:

   ```bash
   kubectl get serviceinstance redis-service -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. Go to the **Catalog** view under the **Service Management** section in the left navigation panel.

2. Switch to the **Add-Ons** tab and select **[Experimental] Redis**.

 > **TIP:** You can also use the search box in the upper right corner of the Console UI to find the service.

3. Select **Add** to provision the Redis ServiceClass and create its instance in the `orders-service` Namespace.

4. Change the **Name** to `redis-service` to match the name of the service, leave `micro` in the **Plan** drop-down list, and set **Image pull policy** to `Always`.

<!-- Explain the image pull policy choice-->

5. Select **Create** to confirm changes.

You will be redirected to **Catalog Management** > **Instances** > **redis-service** view. Wait until the status of the instance changes from `PROVISIONING` to `RUNNING`.

    </details>
</div>

We have a provisioned Redis instance. As the next step, we must bind this instance to the deployed microservice.
