---
title: Bind the Redis ServiceInstance to the microservice
type: Getting Started
---

In this guide, you will bind the created ServiceInstance of the Redis service to the `orders-service` microservice. Then, you will test the binding by sending a sample order to the microservice. This way you can check that the binding to the external Redis storage works and the order data no longer gets removed upon the microservice Pod restart.

## Reference

To create the binding, you will use ServiceBinding and ServiceBindingUsage custom resources (CRs) managed by the [Service Catalog](https://svc-cat.io/docs/walkthrough/).

>**NOTE:** Read more about [provisioning and binding ServiceInstances to microservices in Kyma](/components/service-catalog/#details-provisioning-and-binding).

## Steps

### Bind the Redis ServiceInstance to the microservice

Follow these steps:

<div tabs name="bind-redis-to-microservice" group="bind-redis-to-microservice">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Create a [ServiceBinding CR](https://svc-cat.io/docs/walkthrough/#step-5---requesting-a-servicebinding-to-use-the-serviceinstance) that, in its **spec.instanceRef** field, points to the Redis ServiceInstance created in the previous guide:

   ```yaml
   cat <<EOF | kubectl apply -f -
   apiVersion: servicecatalog.k8s.io/v1beta1
   kind: ServiceBinding
   metadata:
     name: orders-service
     namespace: orders-service
   spec:
     instanceRef:
       name: redis-service
   EOF
   ```

2. Check that the ServiceBinding CR was created. This is indicated by the last condition in the CR status being `Ready True`:

   ```bash
   kubectl get servicebinding orders-service -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

3. Create a [ServiceBindingUsage CR](/components/service-catalog/#custom-resource-service-binding-usage) that injects the Secret associated with the ServiceBinding to the microservice Deployment.

   ```yaml
   cat <<EOF | kubectl apply -f -
   apiVersion: servicecatalog.kyma-project.io/v1alpha1
   kind: ServiceBindingUsage
   metadata:
     name: orders-service
     namespace: orders-service
   spec:
     serviceBindingRef:
       name: orders-service
     usedBy:
       kind: deployment
       name: orders-service
     parameters:
       envPrefix:
         name: "REDIS_"
   EOF
   ```

   - The **spec.serviceBindingRef** and **spec.usedBy** fields are required. **spec.serviceBindingRef** points to the ServiceBinding you have just created and **spec.usedBy** points to the `orders-service` Deployment. More specifically, **spec.usedBy** refers to the name of the Deployment and the cluster-specific [UsageKind CR](/components/service-catalog/#custom-resource-usage-kind) (`kind: deployment`) that defines how Secrets should be injected to `orders-service` microservice when creating a ServiceBinding.

   - The **spec.parameters.envPrefix.name** field is optional. On creating a ServiceBinding, it adds a prefix to all environment variables injected in a Secret to the microservice. In our example, **envPrefix** is `REDIS_`, so all environment variables will follow the `REDIS_{env}` naming pattern.

     > **TIP:** It is considered good practice to use **envPrefix**. In some cases, a microservice must use several instances of a given ServiceClass. Prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another.

4. Check that the ServiceBindingUsage CR was created. This is indicated by the last condition in the CR status being `Ready True`:

   ```bash
   kubectl get servicebindingusage orders-service -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
   ```

 To see the Secret details and retrieve them from the ServiceBinding, run this command:

    ```bash
    kubectl get secret orders-service -n orders-service -o go-template='{{range $k,$v := .data}}{{printf "%s: " $k}}{{if not $v}}{{$v}}{{else}}{{$v | base64decode}}{{end}}{{"\n"}}{{end}}'
    ```

    Expect a response similar to this one:

    ```bash
    HOST: hb-redis-micro-0e965585-9699-443f-b987-38bc6af0e416-redis.orders-service.svc.cluster.local
    PORT: 6379
    REDIS_PASSWORD: 1tvDcINZvp
    ```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. From the `orders-service` Namespace view, go to **Service Management** > **Instances** in the left navigation panel.

2. Switch to the **Add-Ons** tab.

3. Select the `redis-service` item on the list to get into the details view of the `redis-service` Redis instance.

4. Switch to the **Bound Applications** tab.

5. Select **Bind Application**.

6. In the pop-up box that opens up:

    - From the **Select Application** drop-down list, select `order-service`.
    - Select **Set prefix for injected variables** and enter `REDIS_` in the box under the **Prefix namespace value** field.

   > **NOTE:** The **Prefix for injected variables** field is optional. On creating a ServiceBinding, it adds a prefix to all environment variables injected in a Secret to the microservice. In our example, the prefix is set to `REDIS_`, so all environment variables will follow the `REDIS_{ENVIRONMENT_VARIABLE}` naming pattern.

   > **TIP:** It is considered good practice to use prefixes for environment variables. In some cases, a microservice must use several instances of a given ServiceClass. Prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another.

7. Select **Bind Application** to confirm the changes and wait until the status of the created Service Binding Usage changes to `READY`.

  </details>
</div>

### Call and test the microservice

Follow these steps:

> **CAUTION:** If you have a Minikube cluster, you must first add its IP address mapped to the hostname of the exposed Kubernetes Service to the `hosts` file on your machine:
>
>  ```bash
>  echo "$(minikube ip) orders-service.kyma.local" | sudo tee -a /etc/hosts
>  ```

1. Retrieve the domain of the exposed microservice and save it to an environment variable:

   ```bash
   export SERVICE_DOMAIN=$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=orders-service.orders-service -n orders-service -o=jsonpath='{.items[*].spec.hosts[0]}')
   ```

2. Run this command in the terminal to call the service:

   ```bash
   curl -ik "https://$SERVICE_DOMAIN/orders"
   ```

   Expect a response similar to this one:

   ```bash
   content-length: 2
   content-type: application/json;charset=UTF-8
   date: Mon, 13 Jul 2020 13:05:33 GMT
   server: istio-envoy
   vary: Origin
   x-envoy-upstream-service-time: 37

   []
   ```

3. Send a `POST` request to the microservice with sample order details:

   ```bash
   curl -ikX POST "https://$SERVICE_DOMAIN/orders" \
     -H 'Cache-Control: no-cache' -d \
     '{
       "orderCode": "762727210",
       "consignmentCode": "76272725",
       "consignmentStatus": "PICKUP_COMPLETE"
     }'
   ```

4. When you call the `https://$SERVICE_DOMAIN/orders` URL again, the system returns a response with order details similar to these:

   ```bash
   HTTP/2 200
   content-length: 73
   content-type: application/json;charset=UTF-8
   date: Mon, 13 Jul 2020 13:05:51 GMT
   server: istio-envoy
   vary: Origin
   x-envoy-upstream-service-time: 6

   [{"orderCode":"762727210","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```

5. Similarly to the previous guide where you exposed the `orders-service` microservice, remove the Pod created by the `orders-service` Deployment. Run this command, and wait for the system to delete the Pod and start a new one:

   ```bash
   kubectl delete pod -n orders-service -l app=orders-service
   ```

6. Call the microservice again to check the storage:

   ```bash
   curl -ik "https://$SERVICE_DOMAIN/orders"
   ```

   Expect a response similar to this one:

   ```bash
   content-length: 2
   content-type: application/json;charset=UTF-8
   date: Mon, 13 Jul 2020 13:05:33 GMT
   server: istio-envoy
   vary: Origin
   x-envoy-upstream-service-time: 37

   [{"orderCode":"762727210","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```

In the [Expose the microservice](#getting-started-expose-the-microservice) guide, we saw how the in-memory storage works — every time you deleted the Pod of the microservice or changed the Deployment definition, the order details were lost. Thanks to the binding to the Redis instance, order details are now stored outside the microservice and the data is preserved.
