---
title: Configure default ClusterChannelProvisioner
type: Details
---


Kyma comes with NATS Streaming as its default ClusterChannelProvisioner. You can see the configuration details in the [`default-channel-webhook`](../../resources/knative-eventing/charts/knative-eventing/templates/eventing.yaml) ConfigMap. 

You can use a different messaging middleware, other than NATS Streaming, as the Kyma eventing operator. 
To achieve that, configure:
- A ClusterChannelProvisioner that connects with the running messaging middleware 
- The `default-channel-webhook` to use that particular ClusterChannelProvisioner

Read about the examples and the configuration details. 

## In-memory channel
Follow this [guide](https://github.com/knative/eventing/tree/master/config/provisioners/in-memory-channel) to add an in-memory ClusterChannelProvisioner.

>**NOTE**: Before installing this provisioner, add the following annotation to the [`podTemplate.Spec`](https://github.com/knative/eventing/blob/master/config/provisioners/in-memory-channel/in-memory-channel.yaml#L107) in the `in-memory-channel-controller` Deployment to remove the Istio sidecar.

```yaml
template:
  annotations:
    sidecar.istio.io/inject: "false"
    metadata:
      labels: *labels
```

You can change the default cluster channel provisioner by editing the ClusterChannelProvisioner entry in the `default-channel-webhook` ConfigMap. For an example of the in-memory ClusterChannelProvisioner configuration, see [this file](https://github.com/knative/eventing/blob/master/config/400-default-channel-config.yaml).

## Google PubSub

After you complete the [prerequisite steps](https://github.com/knative/eventing/tree/release-0.5/contrib/gcppubsub/config#prerequisites) mentioned in the Knative eventing documentation, follow these steps to configure the Google PubSub ClusterChannelProvisioner:

    > **NOTE:** Skip the last step to install `Knative eventing` as it is pre-installed with Kyma.

1. Deploy the Google PubSub ClusterChannelProvisioner:

    ```bash
    sed "s/REPLACE_WITH_GCP_PROJECT/$PROJECT_ID/" ./assets/gcppubsub.yaml | kubectl apply -f -
    ```

2. In  the `default-channel-webhook` located in the `knative-eventing` Namespace, change the value of the **data.default-channel-config.clusterdefault.name** parameter to `gcp-pubsub`.

    ```bash
    kubectl -n knative-eventing edit configmaps default-channel-webhook
    ```

After the change, the ConfigMap should have the following data:

```yaml
apiVersion: v1
data:
  default-channel-config: |
    clusterdefault:
      apiversion: eventing.knative.dev/v1alpha1
      kind: ClusterChannelProvisioner
      name: gcp-pubsub     #this value has to be changed
kind: ConfigMap
metadata:
  creationTimestamp: "2019-06-05T09:40:17Z"
  name: default-channel-webhook
  namespace: knative-eventing
  resourceVersion: "66671"
  selfLink: /api/v1/namespaces/knative-eventing/configmaps/default-channel-webhook
  uid: edab3828-8775-11e9-b70b-42010a840216
```
