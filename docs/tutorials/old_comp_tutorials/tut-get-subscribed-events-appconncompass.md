---
title: Get subscribed events
type: Tutorials
---

The Event Publisher provides an endpoint for fetching subscribed events for the application. To fetch all of them, make a call:

```bash
curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events/subscribed -k --cert {APP_CERT} --key {APP_CERTS_KEY}
```

A successful call returns a list of all active events for the application.