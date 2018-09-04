---
title: API
type: Details
---

Find the Application Connector API documentation in the included Swagger files.

- See [this file](assets/eventsapi.yaml) for the Events API specification.
- See [this file](assets/metadataapi.yaml) for the Metadata API specification.
- See [this file](assets/connectorapi.yaml) for the Connector API specification.

For convenient viewing, open these files using [this](https://editor.swagger.io/) Swagger editor.

You can acquire the API specification of the Metadata Service for a given version using the following command:
```
curl https://gateway.{CLUSTER_NAME}.kyma.cx/{RE_NAME}/v1/metadataapi.yaml
```

To access the Metadata Service's API specification locally, provide the NodePort of the core-nginx-ingress-controller.

To get the NodePort, run this command:

```
kubectl -n kyma-system get svc core-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}'
```

To access the specification, run:

```
curl https://gateway.kyma.local:{NODE_PORT}/{RE_NAME}/v1/metadataapi.yaml
```

You can acquire the API specification for a given version directly from the Connector Service:
```
curl https://connector-service.{CLUSTER_NAME}.kyma.cx/v1/api.yaml
```
