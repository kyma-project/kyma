---
title: Event Service
type: API
---



The Event Service provides an endpoint for fetching subscribed events for the Application. To fetch all of them, make a call:

```bash
curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events/subscribed -k --cert {APP_CERT} --key {APP_CERTS_KEY}
```
 
 The successful call returns a list of all active events for the Application.
 
>**TIP:** For details on the Event Service API specification, see [this file](./assets/eventsapi.yaml).
