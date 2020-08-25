---
title: Deploy the microservice
type: Getting Started
---

You will now start by deploying a standalone [`orders-service`](https://github.com/kyma-project/examples/blob/master/http-db-service/README.md) microservice on a Kyma cluster. This microservice will act as a link between the external application and the Redis service and we will build the whole end-to-end flow around it.

In this guide you will create:

- Deployment in which you specify the configuration of your microservice
- Kubernetes Service through which your microservice will communicate with other resources on the Kyma cluster

## Steps

### Create a Deployment

Create a [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) that provides the microservice definition and enables you to run it on the cluster. The Deployment uses the `eu.gcr.io/kyma-project/pr/orders-service:PR-162` image. This Docker image exposes the `8080` port on which the related Service is listening.

Follow these steps:

<div tabs name="steps" group="deploy-microservice">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

1. Apply the microservice definition to the `orders-service` Namespace on your cluster:

```bash
kubectl apply -f https://raw.githubusercontent.com/kyma-project/examples/master/orders-service/deployment/orders-service-deployment.yaml
```

2. Check if the Deployment was created. The correct Deployment status should set **readyReplicas** to `1`:

```bash
kubectl get deployment orders-service -n orders-service -o=jsonpath="{.status.readyReplicas}"
```

</details>
<details>
<summary label="ui">
UI
</summary>

1. Create `orders-service-deployment.yaml` on your machine containing [this Deployment definition](https://raw.githubusercontent.com/kyma-project/examples/master/orders-service/deployment/orders-service-deployment.yaml).
2. Once in the `orders-service` Namespace overview, select the **Deploy new resource** button.
3. Browse the `orders-service-deployment.yaml` file and select **Deploy** to confirm the changes.
4. Go to the **Deployments** view under the **Operation** section in the UI to make sure the status of `orders-service` is `RUNNING`.

</details>
</div>

### Create the Service

Deploy the Kubernetes [Service](https://kubernetes.io/docs/concepts/services-networking/service/) in the `orders-service` Namespace to allow other Kubernetes resources to communicate with your microservice.

Follow these steps:

<div tabs name="steps" group="deploy-microservice">
  <details>
  <summary label="kubectl">
  kubectl
  </summary>

Apply the Kubernetes Service to the `orders-service` Namespace on your cluster:

```bash
kubectl apply -f https://raw.githubusercontent.com/kyma-project/examples/master/orders-service/deployment/orders-service-service.yaml
```

  </details>
  <details>
  <summary label="ui">
  UI
  </summary>

1. Create `orders-service-service.yaml` on your machine containing [this Service definition](https://raw.githubusercontent.com/kyma-project/examples/master/orders-service/deployment/orders-service-service.yaml).
2. Once in the `orders-service` Namespace overview, select the **Deploy new resource** button.
3. Browse the `orders-service-service.yaml` file and select **Deploy** to confirm the changes.
4. Go to the **Services** view under the **Operation** section in the UI to make sure the status of `orders-service` is `RUNNING`.

  </details>
  </div>
