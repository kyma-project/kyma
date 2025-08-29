# Create a Workload

This tutorial explains how to create a sample HTTPBin Service Deployment.

## Steps

<!-- tabs:start -->
#### **Kyma Dashboard**

1. Create a namespace with enabled Istio sidecar injection.
2. Go to **Configuration > Service Accounts** and select **Create**. Enter `httpbin` as your Service Account's name and select **Create**.
5. Go to **Discovery and Network > Services** and select **Create**. Provide the following configuration details:
    - **Name**: `httpbin`
    - In the `Labels` section, add the following labels:
      - **service**: `httpbin`
      - **app**:`httpbin`
    - In the `Selectors` section, add the following selector:
      - **app**: `httpbin`
    - In the `Ports` section, select **Add**. Then, use these values:
      - **Name**: `http`
      - **Protocol**: `TCP`
      - **Port**: `8000`
      - **Target Port**: `80`
    - Select **Create**.
8. Go to **Workloads > Deployments** and select **Create**. Choose the HTTPBin template and select **Create**.

#### **kubectl**

1. Create a namespace and export its value as an environment variable. Run:

    ```bash
    export NAMESPACE={NAMESPACE_NAME}
    kubectl create ns $NAMESPACE
    kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
    ```

2. Choose a name for your HTTPBin Service instance and export it as an environment variable.

    ```bash
    export SERVICE_NAME={SERVICE_NAME}
    ```

3. Deploy a sample instance of the HTTPBin Service.

    ```shell
    cat <<EOF | kubectl -n $NAMESPACE apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: $SERVICE_NAME
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: $SERVICE_NAME
      labels:
        app: httpbin
        service: httpbin
    spec:
      ports:
      - name: http
        port: 8000
        targetPort: 80
      selector:
        app: httpbin
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: $SERVICE_NAME
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: httpbin
          version: v1
      template:
        metadata:
          labels:
            app: httpbin
            version: v1
        spec:
          serviceAccountName: $SERVICE_NAME
          containers:
          - image: docker.io/kennethreitz/httpbin
            imagePullPolicy: IfNotPresent
            name: httpbin
            ports:
            - containerPort: 80
    EOF
    ```

4. Verify if an instance of the HTTPBin Service is successfully created.

    ```shell
    kubectl get pods -l app=httpbin -n $NAMESPACE
    ```

    You should get a result similar to this one:

    ```shell
    NAME                        READY    STATUS     RESTARTS    AGE
    {SERVICE_NAME}-{SUFFIX}     2/2      Running    0           96s
    ```
<!-- tabs:end -->