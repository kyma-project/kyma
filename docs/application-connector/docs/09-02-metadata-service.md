---
title: Application Registry
type: API
---

You can get the API specification of the Application Registry for a given version of the service using this command:
```
curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/api.yaml
```

To access the API specification of the Application Registry locally, provide the NodePort of the `application-connector-nginx-ingress-controller`.

To get the NodePort, run this command:

```
kubectl -n kyma-system get svc application-connector-nginx-ingress-controller -o 'jsonpath={.spec.ports[?(@.port==443)].nodePort}'
```

To access the specification, run:

```
curl https://gateway.kyma.local:{NODE_PORT}/{APP_NAME}/v1/metadata/api.yaml
```