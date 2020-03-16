---
title: Set up default Channel
type: Tutorials
---
## Overview

In Knative Eventing Mesh, Channels define an event forwarding and persistence layer. They are responsible for receiving incoming events and forwarding them to resources such as Brokers or other Channels.
By default, Kyma comes with [NatssChannel](https://github.com/knative/eventing-contrib/tree/master/natss/config), but you can change this channel to a different one or even use multiple channels simultaneously.

## Steps
Follow these steps to set up a Channel.

1. Add a ConfigMap with an [override](/root/kyma/#configuration-helm-overrides-for-kyma-installation). For example, this configuration sets the default channel to Kafka channel. 

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
2. Proceed with Kyma installation. If you applied the override in the runtime, trigger the [cluster update process](/root/kyma/#installation-update-kyma).
