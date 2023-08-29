---
title: Register a service
---

This guide shows you how to register a service of your external solution in Kyma. For this example, we use a Basic Authentication-secured API.   

>**NOTE:** Learn how to [register APIs secured with different security schemes or protected against cross-site request forgery (CSRF) attacks](ac-04-register-secured-api.md).

## Prerequisites

1. Before you start, expose the following as environment variables:
   - Your [Application](./ac-01-create-application.md#prerequisites) name
   - Username and password to access the external system
   - Name of the Secret containing the service credentials
   - Name of your service
   - URL to your service
   - Unique ID identifying your service within the Application CR
   - Relative path in your service
   - Namespace in which to create a test Pod

   ```bash
   export APP_NAME=test-app
   export USER_NAME=test-user
   export PASSWORD=test-password
   export SECRET_NAME=test-secret
   export SERVICE_DISPLAY_NAME=test-basic-auth
   export TARGET_URL=https://httpbin.org/
   export TARGET_UUID=f03aafcc-85ad-4665-a46a-bf455f5fa0b3
   export TARGET_PATH=basic-auth/$USER_NAME/$PASSWORD
   export NAMESPACE=default
   ```
     
   > **NOTE:** Replace the example values above with your actual values. 

2. Enable [Istio sidecar injection](/istio/user/00-overview/00-30-overview-istio-sidecars) in the Namespace:
   ```bash
   kubectl label namespace $NAMESPACE istio-injection=enabled
   ```

## Register a service

1. Create a Secret that contains your username and password to the external service:

    ```bash
    kubectl create secret generic $SECRET_NAME --from-literal username=$USER_NAME --from-literal password=$PASSWORD -n kyma-system
    ```

2. To register a service with a Basic Authentication-secured API, you must create or modify the respective Application CR. To create an Application CR with the service definition, run this command:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: applicationconnector.kyma-project.io/v1alpha1
    kind: Application
    metadata:
      name: $APP_NAME
    spec:
      skipVerify: false
      services:
      - id: $TARGET_UUID
        name: $SERVICE_DISPLAY_NAME
        displayName: $SERVICE_DISPLAY_NAME
        description: "Your service"
        providerDisplayName: "Your organisation"
        entries:
        - credentials:
            secretName: $SECRET_NAME
            type: Basic
          targetUrl: $TARGET_URL
          type: API
    EOF
    ```

## Access the registered service 

To check that the service was registered correctly, create a test Pod, and make a call to Application Gateway from within this Pod.   

1. To build a path to access your registered service, run this command:

    ```bash
    export GATEWAY_URL=http://central-application-gateway.kyma-system:8080/$APP_NAME/$SERVICE_DISPLAY_NAME/$TARGET_PATH
    ```
   
    > **CAUTION:** `SERVICE_DISPLAY_NAME` in the **GATEWAY_URL** path must be in its [normalized form](./ac-04-register-secured-api.md#register-a-secured-api). This means that, for example, if you used `test-basic-auth` as the service **displayName**, you're good to go, but if you used `"Test Basic Auth"`, you must replace it with `test-basic-auth` in the path. 

2. Export the name of the test Pod as an environment variable:

    ```bash
    export POD_NAME=test-app-gateway
    ```

3. Create a test Pod:

    ```bash
    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: Pod
    metadata:
      labels:
        run: $POD_NAME
      name: $POD_NAME
      namespace: $NAMESPACE
    spec:
      containers:
      - image: busybox
        name: $POD_NAME
        resources: {}
        tty: true
        stdin: true
      dnsPolicy: ClusterFirst
      restartPolicy: Never
    status: {}
    EOF
    ```

4. Wait for the Pod to be in state `Running`. To check that the Pod is ready, run this command and wait for the response:

    ```bash
    kubectl wait --for=condition=Ready pod $POD_NAME -n $NAMESPACE
    ```

5. To make a call to Application Gateway from within the Pod, run: 

    ```bash
    kubectl exec $POD_NAME -c $POD_NAME -n $NAMESPACE -- sh -c "wget -O- '$GATEWAY_URL'"
    ```

   A successful response from the service means that it was registered correctly.