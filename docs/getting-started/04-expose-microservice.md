---
title: Expose the microservice
type: Getting Started
---

Now that you deployed a standalone `orders-service` microservice, allow other resources to communicate with it. Make it available for other resources outside the cluster by exposing its Kubernetes Service through an APIRule custom resource (CR).

## Reference

This guide demonstrates how [API Gateway](/components/api-gateway) works in Kyma. It exposes your Service on secured or unsecured endpoints for external Services to communicate with it.

## Steps

### Expose the Service

Create an APIRule CR which exposes the Kubernetes Service of the microservice on an unsecured endpoint (**handler** set to `noop`) and accepts the `GET` and `POST` methods.

>**TIP:** **noop** stands for "no operation" and means access without any token. If you want a more secure option, see the [tutorial on how to secure your Service](/components/api-gateway/#tutorials-expose-and-secure-a-service).

Follow these steps:

<div tabs name="steps" group="expose-microservice">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Open the terminal window and apply the [APIRule CR](/components/api-gateway#custom-resource-api-rule):

  ```yaml
  cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: orders-service
    namespace: orders-service
    labels:
      app: orders-service
      example: orders-service
  spec:
    service:
      host: orders-service
      name: orders-service
      port: 80
    gateway: kyma-gateway.kyma-system.svc.cluster.local
    rules:
      - path: /.*
        methods: ["GET","POST"]
        accessStrategies:
          - handler: noop
        mutators: []
  EOF
  ```
2. Check that the API Rule was created and has the status `OK`:

   ```bash
   kubectl get apirules orders-service -n orders-service -o=jsonpath='{.status.APIRuleStatus.code}'
   ```

</details>
<details>
<summary label="console-ui">
Console UI
</summary>

>**TIP:** You can expose a Service or a Function with an API Rule from different views in the Console UI. This tutorial shows how to do that from the generic **API Rules** view.

1. Select the `orders-service` Namespace from the drop-down list in the top navigation panel.

2. Go to **Configuration** > **API Rules** in the left navigation panel and select **Create API Rule**.

3. In the **General settings** section:

    - Enter `orders-service` as the API Rule's **Name**.

    >**NOTE:** The APIRule CR can have a different name than the Service, but it is recommended that all related resources share a common name.

    - Enter `orders-service` as **Hostname** to indicate the host on which you want to expose your Service.

    - From the drop-down list in the **Service** column, select `orders-service (port: 80)` to indicate the Service name for which you want to create the API Rule.

4. In the **Access strategies** section, leave only the `GET` and `POST` methods checked and the `noop` handler selected. This way you will be able to send the orders to the Service and retrieve orders from it without any token.

5. Select **Create** to confirm the changes.

    A message will appear on the screen confirming the changes were saved.

6. In the automatically opened API Rule's details view, check that the API Rule status is `OK`. Then, access the Service by selecting the HTTPS link under **Host** and adding the `/orders` endpoint at the end of it.

</details>
</div>

> **NOTE:** For the whole list of endpoints available in the Service, see its [OpenAPI specification](./assets/orders-service-openapi.yaml).

### Call and test the microservice

Follow these steps:

> **CAUTION:** If you have a Minikube cluster, you must first add its IP address mapped to the hostname of the exposed Kubernetes Service to the `hosts` file on your machine:
>
>  ```bash
>  echo "$(minikube ip) orders-service.kyma.local" | sudo tee -a /etc/hosts
>  ```

1. Retrieve the domain of the exposed microservice and save it to the environment variable:

  ```bash
  export SERVICE_DOMAIN=$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=orders-service.orders-service -n orders-service -o=jsonpath='{.items[*].spec.hosts[0]}')
  ```

2. Run this command in the terminal to call the Service:

  ```bash
  curl -ik "https://$SERVICE_DOMAIN/orders"
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

4. Call the microservice again to check the storage:

  ```bash
  curl -ik "https://$SERVICE_DOMAIN/orders"
  ```

  Expect a response similar to this one:

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

  You can see that the microservice returns the previously sent order details.

5. Remove the [Pod](https://kubernetes.io/docs/concepts/workloads/pods/) created by the `orders-service` Deployment. Run this command, and wait for the system to delete the Pod and start a new one:

   ```bash
   kubectl delete pod -n orders-service -l app=orders-service
   ```

6. Call the microservice again to check the storage:

   ```bash
   curl -ik "https://$SERVICE_DOMAIN/orders"
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

  As you can see, the `orders-service` microservice uses in-memory storage, which means every time you delete the Pod of the microservice or change the Deployment definition, you lose the order details. In further guides, you will see how you can prevent order data deletion by binding external Redis storage to the microservice.
