---
title: Create a ServiceInstance for the Redis service
type: Getting Started
---

To create a binding between the microservice and the Redis service, you must first create an [instance](https://svc-cat.io/docs/walkthrough/#step-4---creating-a-new-serviceinstance) of the previously provisioned service.

## Steps

Follows these steps:

<div tabs name="steps" group="create-redis-instance">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Create a [ServiceInstance CR](https://svc-cat.io/docs/walkthrough/#step-4---creating-a-new-serviceinstance). You will provision the [Redis](https://redis.io/) service with its `micro` plan:

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

2. Check that the ServiceInstance CR was created. This is indicated by the last condition in the CR status being `Ready True`:

   ```bash
   kubectl get serviceinstance redis-service -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. In the left navigation panel, go to **Service Management** > **Catalog**.

2. Switch to the **Add-Ons** tab and select **[Experimental] Redis**.

 > **TIP:** You can also use the search bar in the upper right corner of the Console UI to find the service.

3. Select **Add** to provision Redis and create its instance in the `orders-service` Namespace.

4. Change the **Name** to `redis-service` to match the name of the service, leave `micro` in the **Plan** drop-down list, and set **Image pull policy** to `Always`.

5. Select **Create** to confirm the changes.

You will be redirected to the **Service Management** > **Instances** > **redis-service** view. Wait until the status of the instance changes from `PROVISIONING` to `RUNNING`.

    </details>
</div>

This way you provisioned a Redis instance. As the next step, you will finally fit the two puzzle pieces together and bind the new Redis ServiceInstance to the previously deployed `orders-service` microservice.
