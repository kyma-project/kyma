---
title: Metadata Service
type: API
---

Metadata Service exposes an API built around the concept of service describing external system capabilities.    
Service contains the following:
- API definition comprised of urls and specification following [OpenAPI 2.0 standard](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md).
- Events catalog following [Asynchronous API standard](https://github.com/asyncapi/asyncapi/blob/develop/schema/asyncapi.json).
- Documentation

Service definition may contain both API and Events Catalog definition or only one of those. Providing documentation is optional.  

Metadata Service supports registering APIs secured with OAuth - the user can specify OAuth server url along with client id and client secret.

It is possible to register many services for particular Remote Environment. The Metadata Service API offers a great deal of flexibility as the user may decide how they want to expose their APIs.      

Metadata Service API is described [here](https://github.com/kyma-project/kyma/blob/master/docs/application-connector/docs/assets/metadataapi.yaml).
For a complete information on registering services, please see Managing Registered Services with Metadata API Guide.

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
