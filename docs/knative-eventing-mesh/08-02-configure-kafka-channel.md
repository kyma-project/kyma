---
title: Configure Kafka Channel
type: Tutorials
---

## Overview 

Instead of the defaul Channel implementation, you can use the Knative-compatible [Kafka Channel](https://github.com/kyma-incubator/knative-kafka). This tutorial explains how to configure the connection between Kyma and your Kafka cluster and install Kyma with the Kafka-Channel controller. 

## Steps

Follow these steps to configure the connection between Kyma and the Kafka cluster. 

1. Set up a Kafka cluster on [Azure Event Hub](https://docs.microsoft.com/en-us/azure/event-hubs/event-hubs-create#create-an-event-hub). 

**NOTE**: You can also use other providers, such as [Confluent Cloud](https://www.confluent.io/confluent-cloud) or install Kafka [locally](https://kafka.apache.org/quickstart), but bear in mind that these are experimental options.

2. Export the variables:

```bash
$ export kafkaBrokers={BROKER_URL}
$ export kafkaNamespace={KAFKA_CLUSTER_NAME}
$ export kafkaPassword={PASSWORD}
$ export kafkaUsername={USER_NAME}
$ export kafkaProvider=azure

2. Apply the override:

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

3. Trigger the Kyma installation. You can install Kyma with a `knative-eventing-kafka` custom component by following these [instructions](/root/kyma/#configuration-custom-component-installation).
