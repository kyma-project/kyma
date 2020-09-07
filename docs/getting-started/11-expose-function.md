---
title: Expose the Function
type: Getting Started
---

In this guide, you will expose your Function outside the cluster, through an HTTP proxy. To expose it, use an APIRule custom resource (CR) managed by the in-house API Gateway Controller. This controller reacts to an instance of the APIRule CR and, based on its details, it creates an Istio Virtual Service and Oathkeeper Access Rules that specify your permissions for the exposed Function.

After completing this guide, you will get a Function that:

- Is available on an unsecured endpoint (**handler** set to `noop` in the APIRule CR).
- Accepts the `GET` and `POST` methods.

> **NOTE:** Learn also how to [secure the Function](/components/api-gateway#tutorials-expose-and-secure-a-service-deploy-expose-and-secure-the-sample-resources).

## Steps

### Expose the Function

Follow these steps:

<div tabs name="steps" group="expose-function">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Create an APIRule CR for the Function. It is exposed on port `80`, which is the default port of the [Service](/components/serverless/#architecture-architecture).

  ```yaml
  cat <<EOF | kubectl apply -f -
    apiVersion: gateway.kyma-project.io/v1alpha1
    kind: APIRule
    metadata:
      name: orders-function
      namespace: orders-service
    spec:
      gateway: kyma-gateway.kyma-system.svc.cluster.local
      rules:
        - path: /.*
          accessStrategies:
            - config: {}
              handler: noop
          methods: ["GET","POST"]
      service:
        host: orders-function
        name: orders-function
        port: 80
  EOF
  ```

2. Check that the APIRule was created and has the status `OK`:

  ```bash
  kubectl get apirules orders-function -n orders-service -o=jsonpath='{.status.APIRuleStatus.code}'
  ```

3. Access the Function's external address:

   ```bash
   curl https://orders-function.{CLUSTER_DOMAIN}
   ```

    </details>
    <details>
    <summary label="console-ui">
    Console UI
    </summary>

1. From the drop-down list in the top navigation panel, select the `orders-service` Namespace.

2. Go to **Configuration** > **API Rules** in the left navigation panel and select **Create API Rule**.

3. In the **General settings** section:

    - Enter `orders-function` as the API Rule's **Name**.

    >**NOTE:** The APIRule CR can have a name different from that of the Function, but it is recommended that all related resources share common names.

    - Enter `orders-function` as **Hostname** to indicate the host on which you want to expose your Function.

    - Select `orders-function` as the **Service** that indicates the Function you want to expose.

4. In the **Access strategies** section, leave the default settings, with the `GET` and `POST` methods and the `noop` handler selected.

5. Select **Create** to confirm the changes.

    A message appears on the screen confirming that the changes were saved.

6. Once the pop-up box closes, check that you can access the Function by selecting the HTTPS link under the **Host** column of the new `orders-function` API Rule.

    </details>
</div>

### Call and test the microservice

Follow these steps:

> **CAUTION:** If you have a Minikube cluster, you must first add its IP address mapped to the hostname of the exposed Kubernetes Service to the `hosts` file on your machine:
>
> ```bash
> echo "$(minikube ip) orders-function.kyma.local" | sudo tee -a /etc/hosts
> ```

1. Retrieve the domain of the exposed Function and save it to an environment variable:

   ```bash
   export FUNCTION_DOMAIN=$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=orders-function.orders-service -n orders-service -o=jsonpath='{.items[*].spec.hosts[0]}')
   ```

2. Call the Function:

   ```bash
   curl -ik "https://$FUNCTION_DOMAIN"
   ```

   The system returns a response similar to this one:

   ```bash
   content-length: 2
   content-type: application/json;charset=UTF-8
   date: Mon, 13 Jul 2020 13:05:33 GMT
   server: istio-envoy
   vary: Origin
   x-envoy-upstream-service-time: 37

   []
   ```

3. Send a `POST` request to the Function with sample order details:

   ```bash
   curl -ikX POST "https://$FUNCTION_DOMAIN" \
     -H "Content-Type: application/json" \
     -H 'Cache-Control: no-cache' -d \
     '{
       "orderCode": "762727210",
       "consignmentCode": "76272725",
       "consignmentStatus": "PICKUP_COMPLETE"
     }'
   ```

4. Call the Function again to check the storage:

   ```bash
   curl -ik "https://$FUNCTION_DOMAIN"
   ```

   The system returns a response similar to this one:

   ```bash
   HTTP/2 200
   access-control-allow-origin: *
   content-length: 187
   content-type: application/json; charset=utf-8
   date: Mon, 13 Jul 2020 13:05:51 GMT
   etag: W/"bb-Iuoc/aX9UjdqADJAZPNrwPdoraI"
   server: istio-envoy
   x-envoy-upstream-service-time: 838
   x-powered-by: Express

   [{"orderCode":"762727210","consignmentCode":"76272725","consignmentStatus":"PICKUP_COMPLETE"}]
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
   content-length: 187
   content-type: application/json; charset=utf-8
   date: Mon, 13 Jul 2020 13:05:33 GMT
   etag: W/"bb-Iuoc/aX9UjdqADJAZPNrwPdoraI"
   server: istio-envoy
   x-envoy-upstream-service-time: 838
   x-powered-by: Express

   []
   ```

As you can see, `orders-function` uses in-memory storage, which means every time you delete the Pod of the Function or change its Deployment definition, the order details will be lost. Just like we did with the microservice in the previous guides, let's bind the external Redis storage to the Function to prevent the order data loss.
