---
title: Set up default Channel
type: Tutorials
---

In Knative Eventing Mesh, Channels define an event forwarding and persistence layer. They receive incoming events and dispatch them to resources such as Brokers or other Channels. By default, Kyma comes with [NatssChannel](https://github.com/knative/eventing-contrib/tree/master/natss/config), but you can change it to a different implementation or even use multiple Channels simultaneously. This tutorial shows how to set up Kafka Channel as default.


## Steps
Follow these steps to set up a new default Channel and allow communication between the Channel and the Kafka cluster.

1. Define a ConfigMap with the Kafka Channel [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation). 

```yaml
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
```
2. Create a `yaml` file with an Azure Secret using the specification provided in [this](#tutorials-configure-kafka-channel) tutorial.

3. Use Kyma CLI to install Kyma with the overrides.

  ```bash
  kyma install -o {azure-secret.yaml} -o {kafka-channel.yaml}
  ```