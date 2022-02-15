# E2E Test: External Solution Integration

## Overview

This test verifies if the user can connect an external application to Kyma and interact with its APIs and events using Functions. The test setup mimics a real-world scenario. This means no internal APIs are used and every request sent from the test to the cluster is routed through the ingress-gateway. Additionally, the test application is connected using client certificates, just like an application in a production environment would be connected.

## Test scenario

When you run the test, these actions are performed in the order listed:

1. Create an Application.
2. Create an ApplicationMapping CR for the created Application in the ` e2e-test` Namespace.
3. Deploy a Function in the ` e2e-test` Namespace.
4. Start a test service in the ` e2e-test` Namespace. The Function calls it when it receives an event.
5. Connect the Application through the Application Gateway with client certificates.
6. Register a test service in the Application Registry. The service exposes an event API.
7. Create a ServiceInstance for the registered ServiceClass.
8. Create a ServiceBinding for the ServiceInstance.
9. Create ServiceBindingUsage CR of that binding for the deployed Function.
10. Create a Subscription for the Function, so it is subscribed to the events exposed by the Application.
11. Send an event to the Application Gateway.
12. Verify if the call from the Function reached the test service.

## Compass end-to-end scenario

This scenario uses Compass to register an Application and its APIs and Events.

### Steps

The test performs the following actions:

1. Adds the tested Runtime to the `DEFAULT` scenario in Compass.
2. Registers an Application with API and Event in Compass.
3. Creates the ApplicationMapping CR for the created Application in the `compass-e2e-test` Namespace.
4. Deploys a Function in the `compass-e2e-test` Namespace.
5. Starts a test service in the `compass-e2e-test` Namespace. The Function calls it when it receives an event.
6. Connects the Application through the Application Gateway with client certificates.
7. Creates ServiceInstances for ServiceClasses registered by Compass Runtime Agent (one for API and one for Event services).
8. Creates a ServiceBinding for the API ServiceInstance.
9. Creates the ServiceBindingUsage CR of that binding for the deployed Function.
10. Creates a Subscription for the Function, so that it is subscribed to the events exposed by the Application.
11. Sends an event to the Application Gateway.
12. Verifies if the call from the Function reached the test service.

### Compass connectivity adapter test

This scenario uses Compass and Connectivity Adapter to register an Application and its APIs and Events.
Connectivity Adapter is a component which translates the Application Registry and Application Connector REST API to the Compass GraphQL API.

### Steps

The test performs the following actions:
1. Registers an empty Application in Compass.
2. Creates the ApplicationMapping CR for the created Application in the `connectivity-adapter-e2e` Namespace.
3. Deploys a Function in the `connectivity-adapter-e2e` Namespace.
4. Starts a test service in the `connectivity-adapter-e2e` Namespace. The Function calls it when it receives an event.
5. Connects the Application through the Connectivity Adapter with the client certificates.
6. Registers a service for the Application using Connectivity Adapter.
7. Creates ServiceInstances for ServiceClasses registered by Compass Runtime Agent (one for the API and one for the Event services).
8. Creates a ServiceBinding for the API ServiceInstance.
9. Creates the ServiceBindingUsage CR of that binding for the deployed Function.
10. Creates a Subscription for the Function, so that it is subscribed to the events exposed by the Application.
11. Sends an event to the Application Gateway.
12. Verifies if the call from the Function reached the test service.

## Environment variables for connectivity-adapter-e2e and compass-e2e

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
2. Create an ApplicationMapping CR for the created Application in the ` e2e-mesh-ns` Namespace.
3. Deploy a Function in the `e2e-mesh-ns` Namespace.
4. Start a test service in the ` e2e-mesh-ns` Namespace. The Function calls it when it receives an event.
5. Connect the Application through the Application Gateway with client certificates.
6. Register a test service in the Application Registry. The service exposes an event API.
7. Create a ServiceInstance for the registered ServiceClass.
8. Create a ServiceBinding for the ServiceInstance.
9. Create ServiceBindingUsage CR of that binding for the deployed Function.
10. Create a Knative Trigger for the Function, so it is subscribed to the events exposed by the Application.
11. Send a Cloud event to the Application Gateway at `/events` path.
12. Verify if the call from the Function reached the test service.

## Run the test locally

Currently, the test cannot be executed locally, as it requires access to internal cluster resources. To build and run the test using Octopus, execute:
`make clustertest`

>**Requirements:** 
  > * [kyma-cli](https://github.com/kyma-project/cli)
  > * [ko](https://github.com/google/ko)

>**TIP:** If you are running the test on a cluster with invalid or self-signed SSL certificates, use the `--skipSSLVerify` flag.

### Debug it locally

>**Requirements:** 
  > * [ko](https://github.com/google/ko)
  > * [delve](https://github.com/go-delve/delve)

Even though the test cannot be run locally, there is a way to debug it locally while the test is actually running inside the cluster.
The idea is as follows: 
- build a new container image which contains the delve debugger and the test binary
- deploy a pod which starts the delve debugger and runs the test while waiting for client connection on a specified port
- establish a port-forward between the remote debug server and localhost
- connect with a delve compatible client to the debug server (via port-forward)

#### Start test with delve

Prepare the delve base image. This step needs to be executed once only:

```bash
make build-push-delve-image
```

Resolve dependencies. This step needs to be executed each time the dependencies change:

```bash
make resolve-local
```

Deploy the new test code and start the dlv debugger:

```bash
TEST_DEBUG_SCENARIO=e2e-event-mesh make debug-local
kubectl port-forward -n kyma-system core-test-external-solution 40000
```

#### Connect to the debugger:

**via CLI:**

```shell
dlv connect localhost:40000
```

For further information on how to use dlv, see this [CLI reference](https://github.com/go-delve/delve/tree/master/Documentation/cli).

**via Goland:**

Create a new Run Configuration with type `Go Remote` and select `localhost:40000` as target. Then, click the debug button and debug as usual.

#### Get logs

The test logs can be found inside the container:

```bash
kubectl logs -n kyma-system -l app=core-test-external-solution -c test -f
```
