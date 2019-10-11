---
title: Architecture
---


In Kyma, the tracing component based on Jaeger provides the necessary functionality to collect and query traces. Both operations may occur at the same time, meaning that you can inspect specific traces using the Jeager UI, while at the same time Jaeger takes care of proper trace collection and storage. See the diagram for details: 

![Tracing architecture](./assets/tracing-architecture.svg)


## Collect traces

The process of collecting traces by Jaeger looks as follows:
 
1. The application receives a request, either from an internal or external source.
2. If the application has Istio injection enabled, [Istio proxy](https://github.com/istio/proxy) makes sure to propagate the correct [HTTP headers](/components/tracing#details-propagate-http-headers) of the requests further to the Jaeger Deployment. Istio proxy accesses the Deployment using the [Zipkin](https://zipkin.io/) service which exposes a Jaeger port compatible with the Zipkin protocol.  
3. Jaeger processes the data. Specifically, the  Jaeger Agent component receives the spans, batches them and forwards to the Jaeger Collector service. 
4. The BadgerDB database stores the data and persists it using the PersistentVolume resource.

## Query traces

The process of querying traces from Jaeger looks as follows:

1. A Kyma user accesses the Jaeger UI to [look for specific traces](/components/tracing#details-search-for-traces).
2. Jaeger UI passes the request to the Jaeger Query service. The request goes through the [Istio Ingress Gateway](https://kyma-project.io/docs/components/application-connector/#architecture-application-connector-components-istio-ingress-gateway) which forwards the incoming connections to the service.
3. Jaeger Query passes the request to the [Keycloak Gatekeeper](https://github.com/keycloak/keycloak-gatekeeper) for authorization. The Gatekeeper calls [Dex](https://github.com/dexidp/dex) to authenticate the user and the request, and grants further access if the authentication is successful. 
4. The request goes through Istio proxy which provides further authorization mechanisms to ensure proper communication. 
5. Finally, the functionality provided by the Jaeger Deployment allows you to retrieve trace information. 





