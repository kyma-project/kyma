---
title: "Knative Eventing Mesh (Alpha)"
type: Details
---

## Overview

Knative Eventing Mesh leverages Knative Eventing components to build an eventing mesh that provides event routing and pub/sub capabilities. It abstracts the underlying messaging system and allows you to configure different persistence for each Namespace. Kyma components wire the mesh dynamically for event routing. This way, senders can inject Events into the mesh from multiple source points, and subscribers can receive Events based on filters and their access permissions. The [Knative Broker and Trigger](https://knative.dev/docs/eventing/broker-trigger/) CRDs allow the process of Event publishing and consumption to run smoother, thus significantly improving the overall performance.   

 >**NOTE:** Knative Eventing Mesh is available in the alpha version. Use it only for testing purposes.
 
The new Eventing Mesh runs in parallel with the existing Event Bus. The Kyma Event Bus still supports sending Events to the regular eventing endpoint, while a separate Kyma endpoint handles sending Events to the new Knative Eventing Mesh. 

## Event Flow

### Send Events

The diagram shows you the main stages of the Event flow - from the moment the external Application sends it, up to the point when the lambda function receives it. 

>**NOTE**: The flow assumes you have already added a service instance of an external Application to your Namespace in the Kyma Console and created a lambda with an Event trigger. 

![Sending Events](./assets/knative-event-mesh-send-events.svg)

1. The Application sends Events to the [HTTP source adapter](https://github.com/kyma-project/kyma/tree/master/components/event-sources/adapter/http) which is an HTTP server deployed inside the `kyma-integration` Namespace.  

2. The HTTP source adapter forwards the Events to the default [Knative Broker](https://knative.dev/docs/eventing/broker-trigger).

3. The Knative Broker then delivers Events to the proper lambda function. 

### Subscribe to Events 

In the new Knative Eventing Mesh, you can use Knative Triggers to subscribe to any Events delivered to the Broker located in your Namespace.  

![Subscribe to Events](./assets/knative-event-mesh-subscription.svg)

You can also use expressions that allow the Trigger to filter the incoming Events. For details on setting filters, read the **Trigger filtering** section in [this](https://knative.dev/docs/eventing/broker-trigger/) document. 

## Test Knative Eventing Mesh

To reach the new Eventing Mesh, use an HTTP request with the `/events` path. 
For example, if you have used `gateway.example.cx/v1/events` so far, use `gateway.example.cx/events` to make sure you work with the new Eventing Mesh. 

>**NOTE:** The HTTP source adapter only accepts Events compliant with the [CloudEvents 1.0 specification](https://github.com/cloudevents/spec/blob/v1.0/spec.md).

## Use a custom channel implementation

Kyma has _batteries included_, therefore it comes with a default channel implementation which is Natss.
However, Knative eventing allows to have multiple channel implementations simultaneously.
The default channel implementation can be changed during installation using an installation-override like this:

```bash
$ cat << EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: knative-eventing-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: knative-eventing-init
    kyma-project.io/installation: ""
data:
  eventing.defaultChannel.apiVersion: knativekafka.kyma-project.io/v1alpha1
  eventing.defaultChannel.kind: KafkaChannel
EOF
```

In this example the default channel is set to Kafka.

### Kafka

There is a knative compatible [kafka channel implementation](https://github.com/kyma-incubator/knative-kafka) which can be used for production-ready eventing workloads but is still in alpha state.

# TODO(nachtmaar): how to install kafka: azure, confluent, standalone ... add some links

```bash
$ cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: knative-kafka-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: knative-kafka
    kyma-project.io/installation: ""
type: Opaque    
stringData:
  kafka.brokers: $kafkaBrokers
  kafka.namespace: $kafkaNamespace
  kafka.password: $kafkaPassword
  kafka.username: $kafkaUsername
  kafka.secretName: knative-kafka
  environment.kafkaProvider: azure
```

For details on how to install the kafka custom component see [this](#configuration-custom-component-installation) document.
