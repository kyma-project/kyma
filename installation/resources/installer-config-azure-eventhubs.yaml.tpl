---
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
---
