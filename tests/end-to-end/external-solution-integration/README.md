# E2E Test: External Solution Integration

## Overview

This test verifies if the user can connect an external application to Kyma and use lambdas to interact with it using both API and events. Test behaves like a real-world scenario. This means no internal APIs are used and every request from test to cluster is made through the ingress-gateway. Also, the test application is connected using client certificates as a production application would be.

## Scenario

1. Create an Application
2. Create a ApplicationMapping for that application to test namespace
3. Deploy lambda to test namespace
4. Start test service in test namespace. Lambda should call it upon receiving an event
5. Connect application via application gateway (client certificates) 
6. Register test service in application registry. This service exposes event API.
7. Create ServiceInstance for registered ServiceClass
8. Create ServiceBinding for that instance
9. Create ServiceBindingUsage of that binding for deployed lambda
10. Create Subscription for lambda, so it is subscribed to events exposed by application
11. Send event to application gateway
12. Verify that test service has been called by lambda

## Requirements for running locally

* running kyma cluster

## Run test locally

### Run against Kyma cluster on Minikube 
1. In your `/etc/hosts` find entry with Kyma domains. Add `counter-service.kyma.local` at the end. 
2. Run the test using following command:
    ```
    go run ./cmd/runner e2e
    ```
   
### Run against Kyma cluster in the cloud
Run the test using following command:
```
go run ./cmd/runner e2e --domain {DOMAIN}
```

If you are running the test against cluster with invalid SSL certificates (e.g. self-signed), add `--skipSSLVerify` flag.