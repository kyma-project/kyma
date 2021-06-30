# Eventing Chart

This Helm chart contains all components required for eventing in Kyma:

Components:
- publisher-proxy
- controller
- nats-server

## Publisher Proxy

This component receives legacy and Cloud Event publishing requests from the cluster workloads (microservice or Serverless functions) and redirects them to the Enterprise Messaging Service Cloud Event Gateway. It also fetches a list of subscriptions for a connected application. Click [here](../../components/event-publisher-proxy) for more details.

## Controller

This component manages the internal infrastructure in order to receive an event for all subscriptions handled by NATS or Business Event Bus (BEB)(based on the configuration).

## NATS Server

This component manages NATS server which serves as an eventing platform for Kyma eventing.

## Installation

You can install this Helm chart using either Helm or Kyma CLI.

### Using Helm 3:


```bash
# Install subscriptions.eventing.kyma-project.io CRD
kubectl apply -f resources/cluster-essentials/files/subscriptions.eventing.kyma-project.io.crd.yaml
kubectl apply -f resources/cluster-essentials/files/eventingbackends.eventing.kyma-project.io.crd.yaml
kubectl apply -f resources/cluster-essentials/files/.eventing.kyma-project.io.crd.yaml

$ helm install \
    -n kyma-system \
     eventing .
```
