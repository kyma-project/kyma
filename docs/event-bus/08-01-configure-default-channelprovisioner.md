---
title: Subscription
type: Configure default ClusterChannelProvisioner
---

Kyma comes with NATSS as its default ClusterChannelProvisioner(See [default-channel-webhook](../../resources/knative-eventing/charts/knative-eventing/templates/eventing.yaml) ConfigMap). Other than NATSS, one can use any default channel provisioner of one's choice. In order to add a in-memory-channel provisioner follow this [guide](https://github.com/knative/eventing/tree/master/config/provisioners/in-memory-channel).

The default cluster channel provisioner can be changed by editing the [default-channel-webhook](../../resources/knative-eventing/charts/knative-eventing/templates/eventing.yaml) ConfigMap with the name of the ClusterChannelProvisioner. E.g. for in-memory ClusterChannelProvisioner, see [here](https://github.com/knative/eventing/blob/master/config/400-default-channel-config.yaml.).