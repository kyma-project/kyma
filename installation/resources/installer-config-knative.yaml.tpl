apiVersion: v1
kind: ConfigMap
metadata:
  name: knative-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    feature: knative
data:
  global.ingressgateway.serviceName: "knative-ingressgateway"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: knative-core-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: core
    feature: knative
data:
  gateway.portNamePrefix: "kyma_"
  gateway.selector: "{knative: ingressgateway}"