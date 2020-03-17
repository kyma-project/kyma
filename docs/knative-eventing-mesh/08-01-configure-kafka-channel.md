---
title: Configure Kafka Channel
type: Tutorials
---

## Overview 

Instead of the default Channel implementation, you can use the Knative-compatible [Kafka Channel](https://github.com/kyma-incubator/knative-kafka). To ensure Kafka works properly, you must:

* Install the `knative-eventing-kafka`  component to deploy the Kafka controller
* Set up a Kafka cluster using Azure Event Hubs
* Create a Secret whic the controller uses to communicate with the cluster

**NOTE**: Although Event Hubs and Kafka are nearly identical as concepts, they use different naming convention. To avoid confusion, read [this document](https://docs.microsoft.com/en-us/azure/event-hubs/event-hubs-for-kafka-ecosystem-overview#kafka-and-event-hub-conceptual-mapping).

## Prerequisites

Before Kyma installation, [enable](/root/kyma/#configuration-custom-component-installation) the `knative-eventing-kafka` custom component.

## Steps

Follow these steps:

1. Use the Azure portal to create a [resource group](https://docs.microsoft.com/en-us/azure/event-hubs/event-hubs-create#create-a-resource-group) where your will place your cluster.

2. Create an [Event Hub namespace](https://docs.microsoft.com/en-us/azure/event-hubs/event-hubs-create#create-an-event-hubs-namespace) which is an Event Hub representation of the cluster.

**NOTE**: You can also use other providers, such as [Confluent Cloud](https://www.confluent.io/confluent-cloud) or install Kafka [locally](https://kafka.apache.org/quickstart), but bear in mind that these are experimental options.

3. Export the variables. To retrieve the credentials, go to Azure Portal > **All services** > **Event Hubs** and select your Event Hub. 

```bash
$ export kafkaBrokers={BROKER_URL}
$ export kafkaNamespace={KAFKA_CLUSTER_NAME}
$ export kafkaPassword={PASSWORD}
$ export kafkaUsername={CONNECTION_STRING}
$ export kafkaProvider=azure
```
2. Apply the override which creates the Kafka Secret. 

```bash
$ cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: knative-kafka-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: knative-eventing-kafka
    kyma-project.io/installation: ""
type: Opaque    
stringData:
  kafka.brokers: $kafkaBrokers
  kafka.namespace: $kafkaNamespace
  kafka.password: $kafkaPassword
  kafka.username: $kafkaUsername
  kafka.secretName: knative-kafka
  environment.kafkaProvider: $kafkaProvider
EOF
```

>**NOTE:** For additional values, see [this](https://github.com/kyma-incubator/knative-kafka/blob/master/resources/knative-kafka/values.yaml) file.

4. Proceed with the Kyma installation. Once you configure the Kafka Channel, you can [set it up as a default Channel](#tutorials-set-up-default-channel).