---
title: Deploy the microservice
type: Getting Started
---

Learn how to quickly deploy a standalone [`orders-service`](https://github.com/kyma-project/examples/blob/master/http-db-service/README.md) microservice on a Kyma cluster.

You will create:
- Deployment in which you specify the configuration of the microservice
- Kubernetes Service through which your microservice will communicate with other resources on the Kyma cluster

## Steps

### Create a Deployment

1. Create a [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) that provides the microservice definition and enables you to run it on the cluster. The Deployment uses the `eu.gcr.io/kyma-project/pr/orders-service:PR-162` image. This Docker image exposes the `8080` port on which the related Service is listening.

<div tabs name="steps" group="deploy-microservice">
  <details>
  <summary label="cli">
  CLI
  </summary>

```bash
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orders-service
  namespace: orders-service
  labels:
    app: orders-service
    example: orders-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: orders-service
      example: orders-service
  template:
    metadata:
      labels:
        app: orders-service
        example: orders-service
    spec:
      containers:
        - name: orders-service
          image: "eu.gcr.io/kyma-project/pr/orders-service:PR-162"
          imagePullPolicy: IfNotPresent
          resources:
            limits:
              cpu: 20m
              memory: 32Mi
            requests:
              cpu: 10m
              memory: 16Mi
          env:
            - name: APP_PORT
              value: "8080"
            - name: APP_REDIS_PREFIX
              value: "REDIS_"
EOF
```

2. Check if the Deployment was created. The correct Deployment status should set **readyReplicas** to `1`:

```bash
kubectl get deployment orders-service -n orders-service -o=jsonpath="{.status.readyReplicas}"
```

</details>
<details>
<summary label="console-ui">
Console UI
</summary>

1. Create the `deployment.yaml` file with the Deployment definition:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orders-service
  namespace: orders-service
  labels:
    app: orders-service
    example: orders-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: orders-service
      example: orders-service
  template:
    metadata:
      labels:
        app: orders-service
        example: orders-service
    spec:
      containers:
        - name: orders-service
          image: "eu.gcr.io/kyma-project/pr/orders-service:PR-162"
          imagePullPolicy: IfNotPresent
          resources:
            limits:
              cpu: 20m
              memory: 32Mi
            requests:
              cpu: 10m
              memory: 16Mi
          env:
            - name: APP_PORT
              value: "8080"
            - name: APP_REDIS_PREFIX
              value: "REDIS_"
```

2. Once in the `orders-service` Namespace overview, select the **Deploy new resource** button.
3. Browse the `deployment.yaml` file and select **Deploy** to confirm changes.
4. Go to the **Deployments** view under the **Operation** section in the UI to make sure the status of `orders-service` is `RUNNING`.

</details>
</div>

### Create the Service

Deploy the Kubernetes [Service](https://kubernetes.io/docs/concepts/services-networking/service/) in the `orders-service` Namespace to allow other Kubernetes resources to communicate with your microservice.

<div tabs name="steps" group="deploy-microservice">
  <details>
  <summary label="cli">
  CLI
  </summary>

Run this command:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: orders-service
  namespace: orders-service
  labels:
    app: orders-service
    example: orders-service
spec:
  type: ClusterIP
  ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: 8080
  selector:
    app: orders-service
    example: orders-service
EOF
```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Create a YAML file with the Service definition:

  ```yaml
  apiVersion: v1
  kind: Service
  metadata:
    name: orders-service
    namespace: orders-service
    labels:
      app: orders-service
      example: orders-service
  spec:
    type: ClusterIP
    ports:
      - name: http
        port: 80
        protocol: TCP
        targetPort: 8080
    selector:
      app: orders-service
      example: orders-service
  ```

2. Once in the `orders-service` Namespace overview, select the **Deploy new resource** button.
3. Browse the `service.yaml` file and select **Deploy** to confirm changes.
4. Go to the **Services** view under the **Operation** section in the UI to make sure the status of `orders-service` is `RUNNING`.

  </details>
  </div>
