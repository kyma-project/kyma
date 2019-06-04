---
title: Configure default ClusterChannelProvisioner
type: Tutorials
---

Kyma comes with NATSS as its default ClusterChannelProvisioner(See [default-channel-webhook](../../resources/knative-eventing/charts/knative-eventing/templates/eventing.yaml) ConfigMap). Other than NATSS, one can use any default channel provisioner of one's choice. Following are some provisioners which can be used.

### In-memory-channel
Follow this [guide](https://github.com/knative/eventing/tree/master/config/provisioners/in-memory-channel) to to add an in-memory-channel provisioner.

> **NOTE**: Before installing this provisioner, add the following annotation to the [`podTemplate.Spec`](https://github.com/knative/eventing/blob/master/config/provisioners/in-memory-channel/in-memory-channel.yaml#L107) in the `in-memory-channel-controller` Deployment to remove the Istio sidecar.

```yaml
template:
  annotations:
    sidecar.istio.io/inject: "false"
    metadata:
      labels: *labels
```

The default cluster channel provisioner can be changed by editing the [default-channel-webhook](../../resources/knative-eventing/charts/knative-eventing/templates/eventing.yaml) ConfigMap with the name of the ClusterChannelProvisioner. E.g. for in-memory-channel ClusterChannelProvisioner, see [here](https://github.com/knative/eventing/blob/master/config/400-default-channel-config.yaml).
