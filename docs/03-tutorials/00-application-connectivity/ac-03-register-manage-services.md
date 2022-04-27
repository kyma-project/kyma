---
title: Register a service
---
This guide shows you how to register a service of your external solution in Kyma.

## Prerequisites

- Your [Application name exported](ac-01-create-application.md#prerequisites) as an environment variable

- User name for accessing external system exported as an environment variable

  ```bash
  export USER_NAME={EXTERNAL_SYSTEM_USER_NAME}
  ```

- Password for accessing external system exported as an environment variable

  ```bash
  export PASSWORD={EXTERNAL_SYSTEM_PASSWORD}
  ```

- Secret name containing service credentials exported as an environment variable

  ```bash
  export SECRET_NAME={SECRET_WITH_CREDENTIALS}
  ```

- URL to your service 

  ```bash
  export TARGET_API_URL={EXTERNAL_SYSTEM_URL}
  ```

- Unique ID identifying your service within Application CRD

  ```bash
  export TARGET_API_UUID={UNIQUE_ID}
  ```

## Register a service

1. To register a service with a Basic Authentication-secured API, you must create or modify Application CRD. Run this command to create Application CRD with the service definition.

   >**NOTE:** Follow the [tutorial](ac-04-register-secured-api.md) to learn how to register APIs secured with different security schemes or protected against cross-site request forgery (CSRF) attacks.

   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: applicationconnector.kyma-project.io/v1alpha1
   kind: Application
   metadata:
     name: $APP_NAME
   spec:
     accessLabel: test
     services:
     - description: httpbin.org
       displayName: test-proxy-basic-auth
       entries:
       - accessLabel: test-f03aafcc-85ad-4665-a46a-bf455f5fa0b3
         credentials:
           secretName: $SECRET_NAME
           type: Basic
         targetUrl: $TARGET_API_URL
         type: API
       id: $TARGET_API_UUID
       labels:
         connected-app: test
       longDescription: httpbin.org
       name: test-proxy-basic-auth-690bd
       providerDisplayName: SAP Hybris
     skipVerify: false
   EOF
   ```
   
2. Create a secret contains user name and password

   ```bash
   kubectl create secret generic $SECRET_NAME --from-literal username=$USER_NAME --from-literal password=$PASSWORD 
   ```


## Access registered service 

You can access the registered service from within any workload deployed in Kyma cluster. To verify whether service is registered you can create a test Pod and execute a request to Application Gateway.   

1. Create test pod 

```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: null
  labels:
    run: test-app-gateway
  name: test-app-gateway
spec:
  containers:
  - image: busybox
    name: test-app-gateway
    resources: {}
    tty: true
    stdin: true
  dnsPolicy: ClusterFirst
  restartPolicy: Never
status: {}
EOF
```

2. Follow this template to build path for accessing your registered service

   ```bash
   export GATEWAY_URL=http://central-application-gateway.kyma-system:8080/test/test-proxy-basic-auth/{YOUR SERVICE API PATH}
   ```

3. Execute request 

   ```bash
   kubectl exec test-app-gateway -c test-app-gateway -- sh -c "wget -O- '$GATEWAY_URL'"
   ```

   