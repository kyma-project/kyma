# Eventing Chart

This helm charts contains all components required for eventing in Kyma;

Components:
- event-publisher-proxy

## Event-Publisher-Proxy

This component receives Cloud Event publishing requests from the cluster workloads (microservice or Serverless functions) and redirects it to the Enterprise Messaging Service Cloud Event Gateway. See [here](https://github.com/kyma-project/kyma/tree/master/components/event-publisher-proxy) for more details.
