---
title: Deploy and expose a microservice
---

You already know how to [deploy](02-deploy-expose-function.md#create-a-function) and [expose a Function](02-deploy-expose-function.md#expose-the-function). Let's now do the same with a container microservice.
We'll use the Kyma example [`orders-service`](https://github.com/kyma-project/examples/blob/master/orders-service/README.md) for this.

## Deploy the microservice

First, let's create a Deployment that provides the microservice definition and lets you run it on the cluster.

<div tabs name="Create a microservice Deployment" group="deploy-expose-microservice">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
cat <<EOF | kubectl apply -f -
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: orders-service
    namespace: default
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
            image: "eu.gcr.io/kyma-project/develop/orders-service:e8175c63"
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

To check that the Deployment was created successfully, run:
```bash
kubectl get deployment orders-service -o=jsonpath="{.status.readyReplicas}"
```

The operation was successful if the returned number of **readyReplicas** is `1`.

> **NOTE:** You might need to wait a few seconds for the replica to be ready and return the status.

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. From the left navigation, go to **Deployments**.
2. Click on **Create Deployment +**.
3. Choose the **Advanced** view and provide the following parameters:
    - **Name**: `orders-service`
    - **Labels**: add labels `app` and `example` and set their values to `orders-service`
    - **Containers**: enter Docker image `eu.gcr.io/kyma-project/develop/orders-service:68a58069`
    - Check the **Expose a separate Service** box to create a Service for your Deployment and skip the [next Section](03-deploy-expose-microservice.md#create-the-service)
    - **Port**: `80`
    - **Target Port**: `8080`  

    _Optionally_, to save resources, modify these parameters:
    - **Memory requests**: `10Mi`
    - **Memory limits**: `32Mi`
    - **CPU requests (m)**: `16m`
    - **CPU limits (m)**: `20m`  
  
4. Click **Create**.

The operation was successful if the Pod **Status** for the Deployment is `RUNNING`.
  </details>
</div>

### Create the Service

Now that we have the Deployment, let's deploy the Kubernetes [Service](https://kubernetes.io/docs/concepts/services-networking/service/) to allow other Kubernetes resources to communicate with your microservice.

<div tabs name="Create a Service" group="deploy-expose-microservice">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
cat <<EOF | kubectl apply -f -
  apiVersion: v1
  kind: Service
  metadata:
    name: orders-service
    namespace: default
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

To check that the Service was created successfully, run:

```bash
kubectl get service orders-service -o=jsonpath="{.metadata.uid}"
```

The operation was successful if the command returns the **uid** of your Service.

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

As you've already [created the Kubernetes Service with the microservice](03-deploy-expose-microservice.md#deploy-the-microservice), skip this part.

<!--
//TODO: Functionality not added yet. Check with Hasselhoffs in a while.
If you created the Service at the previous step while creating the Deployment, skip this section. Otherwise, you must now create the Service.

1. From the left navigation, go to **Services**.
2. Click on **Create Service +**.
3. ...

The operation was successful if ... .
--->
  </details>
</div>

## Expose the microservice

We have created the Service. Let's now expose it outside the cluster.

> **CAUTION:** Exposing a workload to the outside world is always a potential security vulnerability, so tread carefully. In a production environment, always [secure the workload](../03-tutorials/00-api-exposure/apix-03-expose-and-secure-service.md) you expose.

To expose our microservice, we must create an [API Rule](../05-technical-reference/00-custom-resources/apix-01-apirule.md) CR for it, just like when we [exposed our Function](02-deploy-expose-function.md#expose-the-function).

<div tabs name="Expose the microservice" group="deploy-expose-microservice">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: orders-service
  namespace: default
  labels:
    app: orders-service
    example: orders-service
spec:
  service:
    host: orders-service.$CLUSTER_DOMAIN
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

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. Using the left navigation, go to **Discovery and Network** > **Services** and select your Service.
2. In your Services's view, click on **Create API Rule +**.
3. Provide the **Name** (`orders-service`) and **Subdomain** (`orders-service`) and click **Create**.

> **NOTE:** Alternatively, from the left navigation go to **Discovery and Network** > **API Rules**, click on **Create API Rule +**, and continue with step 2, selecting the appropriate **Service** from the dropdown menu.
  </details>
</div>

### Verify the microservice exposure

Now let's check that the microservice has been exposed successfully.

<div tabs name="Verify microservice exposure" group="deploy-expose-microservice">
  <details open>
  <summary label="kubectl">
  kubectl
  </summary>

Run:

```bash
curl https://orders-service.$CLUSTER_DOMAIN/orders
```

The operation was successful if the command returns the (possibly empty `[]`) list of orders.

  </details>
  <details>
  <summary label="Kyma Dashboard">
  Kyma Dashboard
  </summary>

1. From your Services's view, get the APIRule's **Host**.

   > **NOTE:** Alternatively, from the left navigation go to **APIRules** and get the **Host** URL from there.

2. Paste this **Host** in your browser and add the `/orders` suffix to the end of it, like this: `{HOST}/orders`. Open it.

The operation was successful if the page shows the (possibly empty `[]`) list of orders.
  </details>
</div>
