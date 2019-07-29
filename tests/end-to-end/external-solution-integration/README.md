# E2E Test: External Solution Integration

## Overview

This test verifies if the user can connect an external application to Kyma and interact with its APIs and events using lambda functions. The test setup mimics a real-world scenario. This means no internal APIs are used and every request sent from the test to the cluster is routed through the ingress-gateway. Additionally, the test application is connected using client certificates, just like an application in a production environment would be connected.

## Test scenario

When you run the test, these actions are performed in the order listed: 

1. Create an Application.
2. Create an ApplicationMapping CR for the created application in the ` e2e-test` Namespace.
3. Deploy lambda to test namespace.
4. Start test service in test namespace. Lambda should call it upon receiving an event.
5. Connect application via application gateway (client certificates).
6. Register test service in application registry. This service exposes event API.
7. Create ServiceInstance for registered ServiceClass.
8. Create ServiceBinding for that instance.
9. Create ServiceBindingUsage of that binding for deployed lambda.
10. Create Subscription for lambda, so it is subscribed to events exposed by application.
11. Send event to application gateway.
12. Verify that test service has been called by lambda.

## Requirements for running locally

* running kyma cluster

## Run test locally

### Run against Kyma cluster on Minikube 
1. In your `/etc/hosts` find entry with Kyma domains. Add `counter-service.kyma.local` at the end. 
2. Run the test using the following command:
    ```
    go run ./cmd/runner e2e
    ```
   
### Run against Kyma cluster in the cloud
Run the test using the following command:
```
go run ./cmd/runner e2e --domain {CLUSTER_DOMAIN}
```

>**TIP:** If you are running the test against cluster with invalid SSL certificates (e.g. self-signed), add `--skipSSLVerify` flag.