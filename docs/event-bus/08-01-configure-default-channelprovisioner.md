---
title: Subscription
type: Configure default ClusterChannelProvisioner
---

Kyma comes with NATSS as its default ClusterChannelProvisioner(See [default-channel-webhook](../../resources/knative-eventing/charts/knative-eventing/templates/eventing.yaml) ConfigMap). Other than NATSS, one can use any default channel provisioner of one's choice. Following are some ClusterChannelProvisioners:

### In-memory

In order to add a in-memory-channel provisioner follow this [guide](https://github.com/knative/eventing/tree/master/config/provisioners/in-memory-channel).

> **Note**: Before installing this provisioner, add the annotation below to get rid of the istio-sidecar [here](https://github.com/knative/eventing/blob/master/config/provisioners/in-memory-channel/in-memory-channel.yaml#L107) for the in-memory-channel-controller.
```yaml
annotations:
  sidecar.istio.io/inject: "false"
```

The default cluster channel provisioner can be changed by editing the [default-channel-webhook](../../resources/knative-eventing/charts/knative-eventing/templates/eventing.yaml) ConfigMap with the name of the ClusterChannelProvisioner. E.g. for in-memory ClusterChannelProvisioner, see [here](https://github.com/knative/eventing/blob/master/config/400-default-channel-config.yaml).