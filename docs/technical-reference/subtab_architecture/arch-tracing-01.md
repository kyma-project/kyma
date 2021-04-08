---
title: Tracing Architecture
type: Architecture
---


The Jaeger-based tracing component provides the necessary functionality to collect and query traces. Both operations may occur at the same time. This way you inspect specific traces using the Jaeger UI, while Jaeger takes care of proper trace collection and storage in parallel. See the diagram for details: 

![Tracing architecture](./assets/tracing-architecture.svg)


## Collect traces

The process of collecting traces by Jaeger looks as follows:
 
1. The application receives a request, either from an internal or external source.
2. If the application has Istio injection enabled, [Istio proxy](https://github.com/istio/proxy) propagates the correct [HTTP headers](/components/tracing#details-propagate-http-headers) of the requests to the Jaeger Deployment. Istio proxy calls Jaeger using the [Zipkin](https://zipkin.io/) service which exposes a Jaeger port compatible with the Zipkin protocol.  
3. Jaeger processes the data. Specifically, the Jaeger Agent component receives the spans, batches them, and forwards to the Jaeger Collector service. 
4. The BadgerDB database stores the data and persists it using a PersistentVolume resource.

## Query traces

The process of querying traces from Jaeger looks as follows:

1. A Kyma user accesses the Jaeger UI to [look for specific traces](/components/tracing#details-search-for-traces).
2. Jaeger UI passes the request to the Jaeger Query service. The request goes through the [Istio Ingress Gateway](/components/application-connector/#architecture-application-connector-components-istio-ingress-gateway) which forwards the incoming connections to the service.
3. Jaeger Query passes the request to the [Keycloak Gatekeeper](https://github.com/keycloak/keycloak-gatekeeper) for authorization. The Gatekeeper calls [Dex](https://github.com/dexidp/dex) to authenticate the user and the request, and grants further access if the authentication is successful. 
4. Finally, the functionality provided by the Jaeger Deployment allows you to retrieve trace information. 





