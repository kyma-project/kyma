---
title: Expose the microservice
type: Getting Started
---

Now that you deployed a standalone `orders-service` microservice, you can make it available outside the cluster to other resources by exposing its Kubernetes Service.

## Prerequisites

Go through the [Deploy the microservice](/#tutorials-deploy-microservice) tutorial to apply the `orders-service` microservice in the `orders-service` Namespace on your cluster.

## Steps

### Expose the Service

Create an APIRule custom resource (CR) which exposes the Kubernetes Service of the microservice under na unsecured endpoint (**handler** set to `noop`) and accepts the `GET` and `POST` methods.

>**TIP:** **noop** stands for "no operation" and means access without any token. If you want to secure your Service, read the [tutorial](/components/api-gateway/#tutorials-expose-and-secure-a-service) to learn how to do that.

<div tabs name="steps" group="expose-microservice">
  <details>
  <summary label="cli">
  CLI
  </summary>

1. Open the terminal window and apply the APIRule CR:

```bash
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
2. Check if the API Rule was created successfully and has the `OK` status:

   ```bash
   kubectl get apirules orders-service -n orders-service -o=jsonpath='{.status.APIRuleStatus.code}'
   ```

</details>
<details>
<summary label="console-ui">
Console UI
</summary>

>**TIP:** You can expose a Service or Function with an API Rule from different views in the Console UI. This tutorial shows how to do that from the generic **API Rules** view.

1. Select the `orders-service` Namespace from the drop-down list in the top navigation panel.

2. Navigate to the **Configuration** section in the left navigation panel, go to the **API Rules** view, and select **Create API Rule**.

3. In the **General settings** section:

    - Enter `orders-service` as the API Rule's **Name**.

    >**NOTE:** The APIRule CR can have a different name than the Service, but it is recommended that all related resources share a common name.

    - Enter `orders-service` as **Hostname** to indicate the host on which you want to expose your Service.

    - Select `orders-service (port: 80)` from the drop-down list in the **Service** column to indicate the Service name for which you want to create the API Rule.

4. In the **Access strategies** section, leave only the `GET` and `POST` methods marked and the `noop` handler selected. This way you will be able to send the orders to a service and retrieve orders from it without any token.

5. Select **Create** to confirm the changes.

    The message appears on the screen confirming the changes were saved.

6. In the API Rule's details view that opens up automatically, check if the API Rule status is `OK`. See if you can access the Service by selecting the HTTPS link under **Host** and adding the `/orders` endpoint at the end of the link.

**NOTE:** For the whole list of endpoints available in the service, see its [OpenAPI specification](./assets/orders-service-openapi.yaml).

</details>
</div>

### Call and test the microservice

> **CAUTION:** If you have a Minikube cluster, you must first add the IP address of the exposed k8s Service to the `hosts` file on your machine:
>
>  ```bash
>  echo "$(minikube ip) orders-service.kyma.local" | sudo tee -a /etc/hosts
>  ```

1. Retrieve the domain of the exposed microservice and save it to the environment variable:

```bash
export SERVICE_DOMAIN=$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=orders-service.orders-service -n orders-service -o=jsonpath='{.items[*].spec.hosts[0]}')
```

2. Run this command in the terminal to call the service:

```bash
   curl -ik "https://$SERVICE_DOMAIN/orders"
```

The system returns a response similar to the following:

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

You should receive a similar response:

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

 You can see the microservice returns the previously sent order details.

5. Remove the [Pod](https://kubernetes.io/docs/concepts/workloads/pods/) created by the `orders-service` Deployment. Run this command and wait for the system to delete the Pod and start a new one:

   ```bash
   kubectl delete pod -n orders-service -l app=orders-service
   ```

6. Call the microservice again to check the storage:

   ```bash
   curl -ik "https://$SERVICE_DOMAIN/orders"
   ```
   You should receive a similar response:

  ```bash
  content-length: 2
  content-type: application/json;charset=UTF-8
  date: Mon, 13 Jul 2020 13:05:33 GMT
  server: istio-envoy
  vary: Origin
  x-envoy-upstream-service-time: 37

  []
  ```

  As you can see, the `orders-service` microservice uses in-memory storage which means every time you delete the Pod of the microservice or change the Deployment definition, the orders details will be lost. In further guides, you will see how you can prevent order data deletion by attaching an external Redis storage to the microservice.
