---
title: Expose a microservice
type: Getting Started
---

Now that you deployed a standalone `http-db-service` application in a cluster, you can make it available outside the cluster to other resources by exposing its Kubernetes Service.

## Prerequisites

Go through the [Deploy a application](/#tutorials-deploy-an-application) tutorial to apply the `http-db-service` application in the `test` Namespace on your cluster.

## Steps

## Expose the Service

Create an APIRule resource which exposes the Kubernetes Service of the application under na unsecured endpoint (**handler** set to `noop`) and accepts the `GET` and `POST` methods.

>**TIP:** If you prefer to secure your Service, read the [tutorial](/components/api-gateway/#tutorials-expose-and-secure-a-service) to learn how to do that.

<div tabs name="steps" group="create-service">
  <details>
  <summary label="cli">
  CLI
  </summary>

Open the terminal window and apply the APIRule:

```
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  labels:
    example: gateway-service
  name: http-db-service
  namespace: test
spec:
  service:
    host: http-db-service
    name: http-db-service
    port: 8017
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
    - path: /.*
      methods: ["GET","POST"]
      accessStrategies:
        - handler: noop
      mutators: []
EOF
```
A successfully deployed APIRule prints this result:

```bash
apirule.gateway.kyma-project.io/http-db-service created
```

</details>
<details>
<summary label="console-ui">
Console UI
</summary>

>**TIP:** You can expose a Service or Function with an API Rule from different views in the Console UI. This tutorial shows how to do that from the generic **API Rules** view.

1. Select the `test` Namespace from the drop-down list in the top navigation panel.

2. Go to the **API Rules** view at the bottom of the left navigation panel and select **Add API Rule**.

3. In the **General settings** section:

    - Enter `http-db-service` as the API Rule's **Name**.

    >**NOTE:** The APIRule CR can have a different name than the Service, but it is recommended that all related resources share a common name.

    - Enter `http-db-service` as **Hostname** to indicate the host on which you want to expose your Service.

    - Select the `http-db-service` Service from the drop-down list in the **Service** column.

4. In the **Access strategies** section, leave only the `GET` and `POST` methods marked and the `noop` handler selected.

5. Select **Create** to confirm changes.

    The message appears on the screen confirming the changes were saved.

6. In the API Rule's details view that opens up automatically, check if you can access the Service by selecting the HTTPS link under **Host**.

>**TIP:** Console UI has a separate **API Rules** view from which you to create APIRules for Services and Functions. Still, you can access this view directly from their corresponding views.

</details>
</div>

## Call and test the Service

> **CAUTION:** If you have a Minikube cluster, you must first add the IP address of your local cluster to the `hosts` file on your machine:

<!-- Improve this caution message to explain exactly why we do that-->

```bash
echo "$(minikube ip) http-db-service.kyma.local" | sudo tee -a /etc/hosts
```

1. Run this command in the terminal to call the service:

```bash
curl -ik "https://$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=http-db-service.test -n test -o=jsonpath='{.items[*].spec.hosts[0]}')/orders"
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

2. Send a `POST` request to the Service with a sample order details:

```bash
curl -ikX POST \
  "https://$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=http-db-service.test -n test -o=jsonpath='{.items[*].spec.hosts[0]}')/orders" \
  -H 'Cache-Control: no-cache' \
  -H 'Content-Type: application/json' \
  -d '{
  "orderId": "118543GU110615ELIN54ZQ",
  "namespace": "test",
  "total": 1234.56
}'
```

3. Run this command in the terminal to call the service:

```bash
curl -ik "https://$(kubectl get virtualservices -l apirule.gateway.kyma-project.io/v1alpha1=http-db-service.test -n test -o=jsonpath='{.items[*].spec.hosts[0]}')/orders"
```

The system returns a response similar to the following:

```bash
HTTP/2 200
content-length: 73
content-type: application/json;charset=UTF-8
date: Mon, 13 Jul 2020 13:05:51 GMT
server: istio-envoy
vary: Origin
x-envoy-upstream-service-time: 6

[{"orderId":"118543GU110615ELIN54ZQ","namespace":"test","total":1234.56}]
```

You can see the service returns the order details previously sent to it.
