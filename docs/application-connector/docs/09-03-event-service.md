---
title: Event Service
type: API
---

See [this file](./assets/eventsapi.yaml) for the Event Service API specification.

###Fetching all active events for the application

Event service provides enpoint for fetching all active events for the application. To do so, make a call:

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
 
 Succesfull call will return a list of all active events for the application.