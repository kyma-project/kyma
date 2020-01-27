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

## Other Channel options

By default Kyma comes with Natss but Knative eventing allows to exchange the default channel implementation. It even allows to have multiple channel implementations simultaneously.
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
    component: knative-eventing
    kyma-project.io/installation: ""
data:
  knative-eventing.channel.default.apiVersion: knativekafka.kyma-project.io/v1alpha1
  knative-eventing.channel.default.kind: KafkaChannel
EOF
```

In this example the default channel is set to Kafka.

### Kafka

There is a Knative compatible [Kafka channel implementation](https://github.com/kyma-incubator/knative-kafka) which can be used for more production-ready eventing workloads but is still in alpha state.

>**NOTE:** Kafka Channel integration is in alpha version. Use it only for testing purposes.

The Knative channel implementation supports [these providers](https://github.com/kyma-incubator/knative-kafka/blob/9eb3fa3f6e67ffc80b162d2ef4c8a8a3942d9c5f/resources/README.md#kafka-providers):

1. [Azure Event Hubs](https://azure.microsoft.com/en-us/services/event-hubs/)
2. [Confluent Cloud](https://www.confluent.io/confluent-cloud)
3. [Standard Kafka installation with no special authorization required](https://kafka.apache.org/quickstart)

Please follow any of the links above on how to setup a Kafka cluster.

Before starting the Kyma installation, the connection between Kyma and the Kafka cluster needs to be configured. This can be done via an installation override like this:

```bash
$ export $kafkaBrokers=<todo user>
$ export $kafkaNamespace=<todo user>
$ export $kafkaPassword=<todo user>
$ export $kafkaUsername=<todo user>
$ export $kafkaProvider=<local|azure|confluent>

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
  environment.kafkaProvider: $kafkaProvider
```

>**NOTE:** For other options, check this [link](https://github.com/kyma-incubator/knative-kafka/blob/master/resources/knative-kafka/values.yaml).

Now that the installation has been customized, the Kyma installation can be triggered. 
You can install Kyma with a custom component (`knative-eventing-channel-kafka` for the component and `knative-eventing-channel-kafka-test` for the TestDefinition) by following these [instructions](/root/kyma/#configuration-custom-component-installation).
