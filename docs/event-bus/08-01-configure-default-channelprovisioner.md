---
title: Configure default ClusterChannelProvisioner
type: Tutorials
---

Kyma comes with NATSS as its default ClusterChannelProvisioner.  You can see the configuration details in the [default-channel-webhook](../../resources/knative-eventing/charts/knative-eventing/templates/eventing.yaml) ConfigMap. You can, however, use other default channel provisioner of your choice. An example can be the in-memory-channel provisioner.

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

You can change the default cluster channel provisioner by editing the ClusterChannelProvisioner entry in the `default-channel-webhook` ConfigMap. For an example of the in-memory-channel ClusterChannelProvisioner configuration, see [this file](https://github.com/knative/eventing/blob/master/config/400-default-channel-config.yaml).
