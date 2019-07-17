# E2E Test: External Solution Integration

## Scenario

1. Create an Application
2. Create a ApplicationMapping for that application to test namespace
3. Deploy lambda to test namespace
4. Start test service in test namespace. Lambda should call it upon receiving an event
5. Connect application via application gateway (client certificates) 
6. Register test service in application registry. This service exposes event API.
7. Create ServiceInstance for registered ServiceClass
8. Create ServiceBinding for that instance
9. Create ServiceBindingUsage of this ServiceInstance for deployed lambda
10. Create Subscription for lambda, so it is subscribed to events exposed by application
11. Send event to application gateway
12. Verify that test service has been called by lambda

## Requirements

* running kyma cluster

## Running

kubectl apply -f deploy.yaml