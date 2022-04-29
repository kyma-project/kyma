---
title: Register a service
---

This guide shows you how to register a service of your external solution in Kyma. For this example, we use a Basic Authentication-secured API.   

>**NOTE:** Learn how to [register APIs secured with different security schemes or protected against cross-site request forgery (CSRF) attacks](ac-04-register-secured-api.md).

## Prerequisites

Before you start, expose the following as environment variables:
- Your [Application](./ac-01-create-application.md#prerequisites) name
- Username and password to access the external system
- Name of the Secret containing the service credentials
- URL to your service
- Unique ID identifying your service within the Application CR
- Relative path in your service

    ```bash
    export APP_NAME=test-app
    export USER_NAME=test-user
    export PASSWORD=test-password
    export SECRET_NAME=test-secret
    export TARGET_URL=https://httpbin.org/
    export TARGET_UUID=f03aafcc-85ad-4665-a46a-bf455f5fa0b3
    export TARGET_PATH=basic-auth/$USER_NAME/$PASSWORD
    ```
> **NOTE:** Replace the example values above with your actual values. 

## Register a service

1. Create a Secret that contains your username and password to the external service:

    ```bash
    kubectl create secret generic $SECRET_NAME --from-literal username=$USER_NAME --from-literal password=$PASSWORD -n kyma-integration
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
        name: test-basic-auth
        displayName: "Test Basic Auth"
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
    export GATEWAY_URL=http://central-application-gateway.kyma-system:8080/$APP_NAME/test-basic-auth/$TARGET_PATH
    ```

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
    kubectl wait --for=condition=Ready pod $POD_NAME
    ```

5. To make a call to Application Gateway from within the Pod, run: 

    ```bash
    kubectl exec $POD_NAME -c $POD_NAME -- sh -c "wget -O- '$GATEWAY_URL'"
    ```

   A successful response indicates that the service was registered correctly.