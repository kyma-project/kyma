---
title: Connector Service
type: API
---

The Connector Service exposes two separate APIs:

- An internal API available in the Kyma cluster used to initiate certificate generation.
- An external API exposed through Ingress used to finalize certificate generation.

Find the specification of both of these APIs [here](./assets/connectorapi.yaml).

Alternatively, get the API specification directly from the Connector Service:
```
https://connector-service.{CLUSTER_DOMAIN}/v1/api.yaml
```
Run this command to access the API specification on a local Kyma deployment:
```
curl https://connector-service.kyma.local/v1/api.yaml
```
