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
kubectl apply -f installation/resources/crds/eventing/subscriptions.eventing.kyma-project.io.crd.yaml
kubectl apply -f installation/resources/crds/eventing/eventingbackends.eventing.kyma-project.io.crd.yaml

$ helm install \
    -n kyma-system \
     eventing .
```

### Using Kyma CLI:

```bash
kyma deploy --source=local --workspace <KYMA_DIR_PATH> --component=eventing
```

To install Eventing with NATS JetStream enabled, run:
```bash
kyma deploy --source=local --value global.jetstream.enabled=true --workspace <KYMA_DIR_PATH> --component=eventing
```

### Configuring NATS JetStream persistence

The persistence used for the stream in the JetStream backend is configurable using the Eventing Helm chart. By default, you can use the `memory` storage type in the evaluation profile and `file` in the production profile. You can customize it further by passing different values to the `kyma deploy` command. For example, to install the production profile with the `memory` storage type of `2Gi` use:

```bash
kyma deploy --profile production \
 --value global.jetstream.enabled=true \
 --value global.jetstream.storage=memory \
 --value eventing.nats.nats.jetstream.memStorage.size=2Gi
```
