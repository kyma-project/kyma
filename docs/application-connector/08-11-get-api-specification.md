---
title: Get the API specification for AC components
type: Tutorials
---

You can get the API specification directly from this website. When you are in the **Application Connector** component, navigate to the **API Consoles** section on the right and choose the desired API to browse its specification, see it on Github, or download it.
If you would rather do this with a command, follow the instructions in this tutorial. 

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

To get the API specification for the Application Registry for a given vresion of the service, run this comand:

```bash
curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/api.yaml
```