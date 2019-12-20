# E2E Test: External Solution Integration

## Overview

This test verifies if the user can connect an external application to Kyma and interact with its APIs and events using lambda functions. The test setup mimics a real-world scenario. This means no internal APIs are used and every request sent from the test to the cluster is routed through the ingress-gateway. Additionally, the test application is connected using client certificates, just like an application in a production environment would be connected.

## Test scenario

When you run the test, these actions are performed in the order listed: 

1. Create an Application.
2. Create an ApplicationMapping CR for the created application in the ` e2e-test` Namespace.
3. Deploy a lambda function in the ` e2e-test` Namespace.
4. Start a test service in the ` e2e-test` Namespace. The lambda function calls it when it receives an event.
5. Connect an application through the Application Gateway with client certificates. 
6. Register a test service in the Application Registry. The service exposes an event API.
7. Create a ServiceInstance for the registered ServiceClass.
8. Create a ServiceBinding for the ServiceInstance.
9. Create ServiceBindingUsage CR of that binding for the deployed lambda function. 
10. Create a Subscription for lambda function, so it is subscribed to the events exposed by the application.
11. Send an event to the Application Gateway. 
12. Verify if the call from the lambda reached the test service.

## Compass E2E Scenario

This scenario uses Compass to register application and it's APIs / Events.

### Steps

When you run the test, these actions are performed in the order listed: 

1. Add runtime that is tested to default scenario in Compass.
2. Register Application with API and Event in Compass.
3. Create an ApplicationMapping CR for the created application in the `compass-e2e-test` Namespace.
4. Deploy a lambda function in the `compass-e2e-test` Namespace.
5. Start a test service in the `compass-e2e-test` Namespace. The lambda function calls it when it receives an event.
6. Connect an application through the Application Gateway with client certificates using Compass. 
7. Create ServiceInstances for ServiceClasses registered by Compass Runtime Agent (one for API and one for Event services).
8. Create a ServiceBinding for the API ServiceInstance.
9. Create ServiceBindingUsage CR of that binding for the deployed lambda function. 
10. Create a Subscription for lambda function, so it is subscribed to the events exposed by the application.
11. Send an event to the Application Gateway. 
12. Verify if the call from the lambda reached the test service.

### Environment variables

Test requires following environment variables:

| Env name | Description |
| --- | --- |
| `DEX_SECRET_NAME` | Name of Kubernetes secret in which dex user credentials are stored | 
| `DEX_SECRET_NAMESPACE` | Namespace of Kubernetes secret in which dex user credentials are stored |
| `DIRECTOR_URL` | Compass Director URL (shouldn't end with `/graphql`) |
| `TENANT` | Compass Tenant ID |
| `RUNTIME_ID` | Compass Runtime ID | 
| `DOMAIN` | Cluster domain | 

## Run the test locally

### Run against Kyma cluster on Minikube 
1. Add an entry to your system's `/etc/hosts` that maps the `counter-service.kyma.local` to `127.0.0.1` 
2. Run the test using the following command:
    ```
    go run ./cmd/runner e2e
    ```
   
### Run against Kyma cluster in the cloud
Run the test using the following command:
```
go run ./cmd/runner e2e --domain {CLUSTER_DOMAIN}
```

>**TIP:** If you are running the test on a cluster with invalid or self-signed SSL certificates, use the `--skipSSLVerify` flag.
