---
title: Event Service
type: API
---



Event service provides an endpoint for fetching subscribed events for the application. To do so, make a call:

```
curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events/subscribed -k --cert {APP_CERT} --key {APP_CERTS_KEY}
```

To get all events locally, provide the NodePort of the `application-connector-ingress-nginx-ingress-controller`.
                  
To get the NodePort, run this command:
                  
 ```
 kubectl -n kyma-system get svc application-connector-ingress-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}'
 ```
 
 The next step is to run this command:
 ```
 curl https://gateway.kyma.local:{NODE_PORT}/{APP_NAME}/v1/events/subscribed -k --cert {APP_CERT} --key {APP_CERTS_KEY}
 ```
 
 The successful call returns a list of all active Events for the application.
 
>**TIP:** For details on the Event Service API specification, see [this file](./assets/eventsapi.yaml).
