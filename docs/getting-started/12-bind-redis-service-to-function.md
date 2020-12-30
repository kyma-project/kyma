---
title: Bind a Redis ServiceInstance to the Function
type: Getting Started
---

This guide shows how you can bind a sample instance of the Redis service to the Function. After completing all the steps, you will get the Function with encoded Secrets to the service. You will use Redis as an external database for the Function. This database replaces the Function's in-memory storage.

## Steps

### Bind the Redis ServiceInstance to the Function

Follow these steps:

<div tabs name="steps" group="bind-redis-to-function">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Create a ServiceBinding custom resource (CR) that points to the existing Redis instance in the **spec.instanceRef** field:

  ```yaml
  cat <<EOF | kubectl apply -f -
  apiVersion: servicecatalog.k8s.io/v1beta1
  kind: ServiceBinding
  metadata:
    name: orders-function
    namespace: orders-service
  spec:
    instanceRef:
      name: redis-service
  EOF
  ```

2. Check that the ServiceBinding CR was created. This is indicated by the last condition in the CR status equal to `Ready True`:

  ```bash
  kubectl get servicebinding orders-function -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
  ```

3. Create a ServiceBindingUsage CR:

  ```yaml
  cat <<EOF | kubectl apply -f -
  apiVersion: servicecatalog.kyma-project.io/v1alpha1
  kind: ServiceBindingUsage
  metadata:
    name: orders-function
    namespace: orders-service
  spec:
    serviceBindingRef:
      name: orders-function
    usedBy:
      kind: serverless-function
      name: orders-function
    parameters:
      envPrefix:
        name: "REDIS_"
  EOF
  ```

   - The **spec.serviceBindingRef** and **spec.usedBy** fields are required. **spec.serviceBindingRef** points to the ServiceBinding you have just created and **spec.usedBy** points to the Function. More specifically, **spec.usedBy** refers to the name of the Function and the cluster-specific [UsageKind CR](/components/service-catalog/#custom-resource-usage-kind) (`kind: serverless-function`) that defines how Secrets should be injected to your Function when creating a ServiceBinding.

   - The **spec.parameters.envPrefix.name** field is optional. On creating a ServiceBinding, it adds a prefix to all environment variables injected in a Secret to the Function. In our example, **envPrefix** is `REDIS_`, so all environment variables will follow the `REDIS_{env}` naming pattern.

     > **TIP:** It is considered good practice to use **envPrefix**. In some cases, a Function must use several instances of a given ServiceClass. Prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another.

4. Check that the ServiceBindingUsage CR was created. This is indicated by the last condition in the CR status equal to `Ready True`:

  ```bash
  kubectl get servicebindingusage orders-function -n orders-service -o=jsonpath="{range .status.conditions[*]}{.type}{'\t'}{.status}{'\n'}{end}"
  ```

To see the Secret details and retrieve them from the ServiceBinding, run this command:

  ```bash
  kubectl get secret orders-function -n orders-service -o go-template='{{range $k,$v := .data}}{{printf "%s: " $k}}{{if not $v}}{{$v}}{{else}}{{$v | base64decode}}{{end}}{{"\n"}}{{end}}'
  ```

  Expect a response similar to this one:

  ```bash
  HOST: hb-redis-micro-0e965585-9699-443f-b987-38bc6af0e416-redis.serverless.svc.cluster.local
  PORT: 6379
  REDIS_PASSWORD: 1tvDcINZvp
  ```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Go to **Workloads** > **Functions** in the left navigation panel and select `orders-function`.

2. Switch to the **Configuration** tab and select **Create Service Binding** in the **Service Bindings** section.

3. Select `redis-service` from the **Service Instance** drop-down list, add `REDIS_` as **Prefix for injected variables**, and make sure **Create new Secret** is selected.

   > **NOTE:** The **Prefix for injected variables** field is optional. On creating a ServiceBinding, it adds a prefix to all environment variables injected in a Secret to the Function. In our example, the prefix is set to `REDIS_`, so all environment variables will follow the `REDIS_{ENVIRONMENT_VARIABLE}` naming pattern.

   > **TIP:** It is considered good practice to use prefixes for environment variables. In some cases, a Function must use several instances of a given ServiceClass. Prefixes allow you to distinguish between instances and make sure that one Secret does not overwrite another.

4. Select **Create** to confirm the changes.

A message confirming that the ServiceBinding was created will appear in the **Service Bindings** section in your Function's view along with environment variable names.

When you switch to the **Code** tab and scroll down to the **Environment Variables** section, you can now see `REDIS_PORT`, `REDIS_HOST`, and `REDIS_REDIS_PASSWORD` items with the `Service Binding` type. It indicates that the environment variable was injected to the Function by the ServiceBinding.

    </details>
</div>

### Call and test the Function

Follow these steps:

> **CAUTION:** If you have a Minikube cluster, you must first add its IP address mapped to the hostname of the exposed Kubernetes Service to the `hosts` file on your machine:
>
>  ```bash
>  echo "$(minikube ip) orders-service.kyma.local" | sudo tee -a /etc/hosts
>  ```

1. Retrieve the domain of the exposed Function and save it to an environment variable:

   ```bash
   export FUNCTION_DOMAIN=$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=orders-function.orders-service -n orders-service -o=jsonpath='{.items[*].spec.hosts[0]}')
   ```

2. Run this command in the terminal to call the Function:

   ```bash
   curl -ik "https://$FUNCTION_DOMAIN"
   ```

   The system returns a response similar to this one:

   ```bash
   HTTP/2 200
   access-control-allow-origin: *
   content-length: 187
   content-type: application/json; charset=utf-8
   date: Mon, 13 Jul 2020 13:05:33 GMT
   etag: W/"bb-Iuoc/aX9UjdqADJAZPNrwPdoraI"
   server: istio-envoy
   x-envoy-upstream-service-time: 25
   x-powered-by: Express

   [{"orderCode":"762727210","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"123456789","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```

   > **NOTE:** The listed orders are the orders created in the previous guides.

3. Send a `POST` request to the Function with sample order details:

   ```bash
   curl -ikX POST "https://$FUNCTION_DOMAIN" \
     -H "Content-Type: application/json" \
     -H 'Cache-Control: no-cache' -d \
     '{
       "orderCode": "762727234",
       "consignmentCode": "76272725",
       "consignmentStatus": "PICKUP_COMPLETE"
     }'
   ```

4. When you call the `https://$FUNCTION_DOMAIN` URL again, the system returns a response similar to this one:

   ```bash
   HTTP/2 200
   access-control-allow-origin: *
   content-length: 280
   content-type: application/json; charset=utf-8
   date: Mon, 13 Jul 2020 13:05:51 GMT
   etag: W/"118-lAc/HJqUaWFKlQ/uyvrhuPb++80"
   server: istio-envoy
   x-envoy-upstream-service-time: 9
   x-powered-by: Express

   [{"orderCode":"762727234","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"762727210","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"123456789","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```

   You can see the Function returns the previously sent order details.

5. Remove the [Pod](https://kubernetes.io/docs/concepts/workloads/pods/) created by the `orders-function` Function. Run this command, and wait for the system to delete the Pod and start a new one:

   ```bash
   kubectl delete pod -n orders-service -l "serverless.kyma-project.io/function-name=orders-function"
   ```

6. Call the Function again to check the storage:

   ```bash
   curl -ik "https://$FUNCTION_DOMAIN"
   ```

   The system returns a response similar to this one:

   ```bash
   HTTP/2 200
   access-control-allow-origin: *
   content-length: 280
   content-type: application/json; charset=utf-8
   date: Mon, 13 Jul 2020 13:05:33 GMT
   etag: W/"118-lAc/HJqUaWFKlQ/uyvrhuPb++80"
   server: istio-envoy
   x-envoy-upstream-service-time: 975
   x-powered-by: Express

   [{"orderCode":"762727234","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"762727210","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}, {"orderCode":"123456789","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
   ```

As you can see, `orders-function` uses Redis storage now. It means that every time you delete the Pod of the Function or change its Deployment definition, the order details are preserved.
