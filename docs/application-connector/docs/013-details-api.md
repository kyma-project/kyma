---
title: API
type: Details
---

Find the Application Connector API documentation in the included Swagger files.

- See [this file](https://github.com/kyma-project/kyma/blob/master/docs/application-connector/docs/assets/eventsapi.yaml) for the Events API specification.
- See [this file](https://github.com/kyma-project/kyma/blob/master/docs/application-connector/docs/assets/metadataapi.yaml) for the Metadata API specification.
- See [this file](https://github.com/kyma-project/kyma/blob/master/docs/application-connector/docs/assets/connectorapi.yaml) for the Connector API specification.


You can acquire the API specification of the Metadata Service for a given version using the following command:
```
curl https://gateway.{CLUSTER_NAME}.kyma.cx/{RE_NAME}/v1/metadata/api.yaml
```

To access the Metadata Service's API specification locally, provide the NodePort of the `core-nginx-ingress-controller`.

To get the NodePort, run this command:

```
kubectl -n kyma-system get svc core-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}'
```

To access the specification, run:

```
curl https://gateway.kyma.local:{NODE_PORT}/{RE_NAME}/v1/metadata/api.yaml
```
