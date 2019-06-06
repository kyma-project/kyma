---
title: Event Service
type: API
---



The Event Service provides an endpoint for fetching subscribed Events for the application. To fetch all of them, make a call:

```
curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events/subscribed -k --cert {APP_CERT} --key {APP_CERTS_KEY}
```
 
 The successful call returns a list of all active Events for the application.
 
>**TIP:** For details on the Event Service API specification, see [this file](./assets/eventsapi.yaml).
