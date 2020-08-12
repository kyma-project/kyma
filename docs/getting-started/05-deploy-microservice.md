---
title: Deploy a microservice
type: Getting Started
---

Learn how to quickly deploy a standalone [`http-db-service`](https://github.com/kyma-project/examples/blob/master/http-db-service/README.md) microservice on a Kyma cluster.

You will create:
- `test` Namespace for your application
- Deployment in which you specify the application configuration
- Kubernetes Service through which your application will communicate with other resources on the Kyma cluster

## Prerequisites

To use the Kyma cluster and install the example, download these tools:

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.16
- [curl](https://github.com/curl/curl)

## Steps

### Get the kubeconfig file and configure the CLI

Follow these steps to get the `kubeconfig` file and configure the CLI to connect to the cluster:

1. Access the Console UI of your Kyma cluster.
2. Click the user icon in the top right corner.
3. Select **Get Kubeconfig** from the drop-down menu to download the configuration file to a selected location on your machine.
4. Open a terminal window.
5. Export the **KUBECONFIG** environment variable to point to the downloaded `kubeconfig`. Run this command:

   ```bash
   export KUBECONFIG={KUBECONFIG_FILE_PATH}
   ```

   >**NOTE:** Drag and drop the `kubeconfig` file in the terminal to easily add the path of the file to the `export KUBECONFIG` command you run.

6. Run `kubectl cluster-info` to check if the CLI is connected to the correct cluster.

### Create a Namespace

Create the `test` Namespace where you will deploy your Service.

<div tabs name="steps" group="create-service">
  <details>
  <summary label="cli">
  CLI
  </summary>

Run this command:

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: test
EOF
```

</details>
<details>
<summary label="console-ui">
Console UI
</summary>

1. In the main view of the Console UI, select the **Add new namespace** button.
2. Enter `test` in the **Name** field and confirm by selecting the **Create** button.

</details>
</div>

### Create a Deployment

Create a [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) that provides the application definition and enables you to run it on the cluster. The Deployment uses the `eu.gcr.io/kyma-project/develop/http-db-service:47d43e19` image. This Docker image exposes the `8017` port on which the related Service is listening.

<div tabs name="steps" group="create-service">
  <details>
  <summary label="cli">
  CLI
  </summary>

```
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: http-db-service
  namespace: test
  labels:
    example: http-db-service
    app: http-db-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: http-db-service
      example: http-db-service
  template:
    metadata:
      labels:
        app: http-db-service
        example: http-db-service
    spec:
      containers:
      # replace the repository URL with your own repository (e.g. {DockerID}/http-db-service:0.0.x for Docker Hub).
      - image: eu.gcr.io/kyma-project/develop/http-db-service:47d43e19
        imagePullPolicy: IfNotPresent
        name: http-db-service
        ports:
        - name: http
          containerPort: 8017
        resources:
          limits:
            memory: 100Mi
          requests:
            memory: 32Mi
        env:
        - name: dbtype
          # available dbtypes are: [memory, mssql]
          value: "memory"
EOF
```

A successfully created Deployment prints this result:

```bash
deployment.apps/http-db-service created
```

</details>
<details>
<summary label="console-ui">
Console UI
</summary>

1. Create a YAML file with the Deployment definition:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: http-db-service
  namespace: test
  labels:
    example: http-db-service
    app: http-db-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: http-db-service
      example: http-db-service
  template:
    metadata:
      labels:
        app: http-db-service
        example: http-db-service
    spec:
      containers:
      # replace the repository URL with your own repository (e.g. {DockerID}/http-db-service:0.0.x for Docker Hub).
      - image: eu.gcr.io/kyma-project/develop/http-db-service:47d43e19
        imagePullPolicy: IfNotPresent
        name: http-db-service
        ports:
        - name: http
          containerPort: 8017
        resources:
          limits:
            memory: 100Mi
          requests:
            memory: 32Mi
        env:
        - name: dbtype
          # available dbtypes are: [memory, mssql]
          value: "memory"
```

2. Go to the `test` Namespace view in the Console UI and select the **Deploy new resource** button.
3. Browse your Deployment file and select **Deploy** to confirm changes.
4. Go to the **Deployments** view to make sure `http-db-service` is running.

</details>
</div>

### Create the Service

Deploy the Kubernetes `http-db-service` [Service](https://kubernetes.io/docs/concepts/services-networking/service/) in the `test` Namespace to allow other Kubernetes resources to communicate with your application.

<div tabs name="steps" group="create-service">
  <details>
  <summary label="cli">
  CLI
  </summary>

Run this command:

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: http-db-service
  namespace: test
  labels:
    example: http-db-service
    app: http-db-service
spec:
  ports:
  - name: http
    port: 8017
  selector:
    app: http-db-service
    example: http-db-service
EOF
```

A successfully deployed Service prints this result:

```bash
service/http-db-service created
```

  </details>
  <details>
  <summary label="console-ui">
  Console UI
  </summary>

1. Create a YAML file with the Service definition:

  ```
  apiVersion: v1
  kind: Service
  metadata:
    name: http-db-service
    namespace: test
    labels:
      example: http-db-service
      app: http-db-service
  spec:
    ports:
    - name: http
      port: 8017
    selector:
      app: http-db-service
      example: http-db-service
  ```

2. Go to the `test` Namespace view in the Console UI and select the **Deploy new resource** button.
3. Browse your Service file and select **Deploy** to confirm changes.
4. Go to the **Services** view to make sure `http-db-service` is running.

  </details>
  </div>
