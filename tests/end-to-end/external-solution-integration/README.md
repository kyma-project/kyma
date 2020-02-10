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

## Compass end-to-end scenario

This scenario uses Compass to register an Application and its APIs and Events.

### Steps

The test performs the following actions:

1. Adds the tested Runtime to the `DEFAULT` scenario in Compass.
2. Registers an Application with API and Event in Compass.
3. Creates the ApplicationMapping CR for the created Application in the `compass-e2e-test` Namespace.
4. Deploys a lambda function in the `compass-e2e-test` Namespace.
5. Starts a test service in the `compass-e2e-test` Namespace. The lambda function calls it when it receives an event.
6. Connects the Application through the Application Gateway with client certificates. 
7. Creates ServiceInstances for ServiceClasses registered by Compass Runtime Agent (one for API and one for Event services).
8. Creates a ServiceBinding for the API ServiceInstance.
9. Creates the ServiceBindingUsage CR of that binding for the deployed lambda function. 
10. Creates a Subscription for the lambda function, so that it is subscribed to the events exposed by the Application.
11. Sends an event to the Application Gateway. 
12. Verifies if the call from the lambda reached the test service.

### Compass connectivity adapter test

This scenario uses Compass and Connectivity Adapter to register an Application and its APIs and Events.
Connectivity Adapter is a component which translates Application Registry and Application Connector REST API to Compass GraphQL API. 

### Steps

The test performs the following actions:
1. Registers an empty Application in Compass.
2. Creates the ApplicationMapping CR for the created Application in the `connectivity-adapter-e2e` Namespace.
3. Deploys a lambda function in the `connectivity-adapter-e2e` Namespace.
4. Starts a test service in the `connectivity-adapter-e2e` Namespace. The lambda function calls it when it receives an event.
6. Connects the Application through the Connectivity Adapter with the client certificates. 
5. Registers Service for the Application using Connectivity Adapter
7. Creates ServiceInstances for ServiceClasses registered by Compass Runtime Agent (one for the API and one for the Event services).
8. Creates a ServiceBinding for the API ServiceInstance.
9. Creates the ServiceBindingUsage CR of that binding for the deployed lambda function. 
10. Creates a Subscription for the lambda function, so that it is subscribed to the events exposed by the Application.
11. Sends an event to the Application Gateway. 
12. Verifies if the call from the lambda reached the test service.

## Environment variables for connectivity-adapter-e2e and compass e2e

The test requires the following environment variables:

| Environment variable name | Description |
| --- | --- |
| `DEX_SECRET_NAME` | Name of the Kubernetes Secret which stores Dex user credentials | 
| `DEX_SECRET_NAMESPACE` | Namespace of the Kubernetes Secret which stores Dex user credentials |
| `DIRECTOR_URL` | Compass Director URL. The URL should not end with `/graphql`. |
| `TENANT` | Compass Tenant ID |
| `RUNTIME_ID` | Compass Runtime ID | 
| `DOMAIN` | Cluster domain |

## Test scenario (with Event Mesh Alpha)

When you run the test, these actions are performed in the order listed: 

1. Create an Application.
2. Create an ApplicationMapping CR for the created application in the ` e2e-mesh-ns` Namespace.
3. Deploy a lambda function in the `e2e-mesh-ns` Namespace.
4. Start a test service in the ` e2e-mesh-ns` Namespace. The lambda function calls it when it receives an event.
5. Connect an application through the Application Gateway with client certificates. 
6. Register a test service in the Application Registry. The service exposes an event API.
7. Create a ServiceInstance for the registered ServiceClass.
8. Create a ServiceBinding for the ServiceInstance.
9. Create ServiceBindingUsage CR of that binding for the deployed lambda function. 
10. Create a Knative Trigger for lambda function, so it is subscribed to the events exposed by the application.
11. Send a Cloud event to the Application Gateway at `/events` path. 
12. Verify if the call from the lambda reached the test service.

## Run the test locally

### Run against Kyma cluster on Minikube 
1. Add an entry to your system's `/etc/hosts` that maps the `counter-service-{name-of-test}.kyma.local` to `minikube cluster IP`,
e.g.: for *compass-e2e*, it will be counter-service-compass-e2e-test.kyma.local 
2. Set required ENVs
3. Run the test using the following command:
    ```
    go run ./cmd/runner {name-of-test}
    ```
   
### Run against Kyma cluster in the cloud
Run the test using the following command:
```
go run ./cmd/runner e2e --domain {CLUSTER_DOMAIN}
```

>**TIP:** If you are running the test on a cluster with invalid or self-signed SSL certificates, use the `--skipSSLVerify` flag.
