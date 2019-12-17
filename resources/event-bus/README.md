# Event Bus

## Prerequisites

This sub-chart requires:
* Kubernetes 1.9
* The mutual Transport Layer Security (TLS) between clients
* The Event Bus deployments requires:
  * A Kubernetes cluster with Istio installed
  * The istio-side-car injection enabled

For more details, see the [Istio documentation](https://istio.io/docs/).

## Details
Configure these options for each business requirement:

|Component                 | Configuration  | Description |
|:---------------------------|:--------|:-----------|
| **nats-streaming** |
| | `global.persistence.size` | The size of the storage volume. |
| | `global.persistence.maxAge`| The maximum period of time for storing an event (`0` for unlimited). |
| |`global.natsStreaming.resources`| Refer to Kubernetes resource requests and limits for details. |
| |`global.natsStreaming.channel.maxInactivity`| The maximum inactivity period (without any new Event or subscription) after which a channel can be garbage collected (`0` for unlimited). |
| **eventPublishService** |
| |`global.eventPublishService.maxRequests`| The maximum number of concurrent events to publish. |
| |`global.eventPublishService.resources`| Refer to Kubernetes resource requests and limits for details. |


For details on the Kubernetes resource requests and limits, see the [Manage Compute Resources Container](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) document.
