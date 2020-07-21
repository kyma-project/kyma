---
title: Get the API specification for AC components
type: Tutorials
---

To view or download the API specification directly from this website, see the **API Consoles** section. Alternatively, get it by calling the API directly. To do so, follow the instructions in this tutorial. 

## Connector Service API

To get the API specification for the Connector Service on a local Kyma deployment, run this command:

```bash
curl https://connector-service.kyma.local/v1/api.yaml
```

Alternatively, get the API specification directly from the Connector Service: 

```bash
https://connector-service.{CLUSTER_DOMAIN}/v1/api.yaml
```

## Application Registry API

To get the API specification for the Application Registry for a given version of the service, run this command:

```bash
curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/api.yaml
```