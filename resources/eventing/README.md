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
kyma deploy --source=local --value global.features.enableJetstream=true --workspace <KYMA_DIR_PATH> --component=eventing
```

### Configuring NATS JetStream persistence

The persistence used for the stream used in the JetStream Backend is configurable via the Eventing controller Helm chart. By default, the Evaluation profile uses storage type `memory` and the Productionn profile uses the `file` storage type. This can be further customized by passing different values to the `kyma deploy` command. For example, to install the Production profile with `memory` storage type of `2Gi`:

```bash
kyma deploy --profile production \
 --value global.features.enableJetstream=true \
 --value eventing.controller.jetstream.storage=memory \
 --value eventing.nats.nats.jetstream.memStorage.size=2Gi
```
