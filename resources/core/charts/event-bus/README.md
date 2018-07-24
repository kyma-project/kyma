```
  ______               _     ____            
 |  ____|             | |   |  _ \           
 | |____   _____ _ __ | |_  | |_) |_   _ ___
 |  __\ \ / / _ \ '_ \| __| |  _ <| | | / __|
 | |___\ V /  __/ | | | |_  | |_) | |_| \__ \
 |______\_/ \___|_| |_|\__| |____/ \__,_|___/
```

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

| Component                 | Configuration  | Description |
|---------------------------| --------:| -----------: |
| **nats-streaming** |
| | `global.persistence.size` | The size of the storage volume. |
| | `global.persistence.maxAge`| The maximum period of time for storing an event (`0` for unlimited). |
| |`global.natsStreaming.resources`| Refer to Kubernetes resource requests and limits for details. |
| **publish** |
| |`global.publish.maxRequests`| The maximum number of concurrent events to publish. |
| |`global.publish.resources`| Refer to Kubernetes resource requests and limits for details. |
| **push**|
| | `global.push.http.subscriptionNameHeader` | The HTTP header that contains the push subscription name. |
| | `global.push.http.topicHeader` | The HTTP header that contains the `event-type` details. |
| |`global.push.resources`| Refer to Kubernetes resource requests and limits for details. |
| **sub-validator** |
| | `global.subValidator.resyncPeriod`| The period after which the synchronization of EventActivation and Subscription Kubernetes custom resources takes place. |
| |`global.subValidator.resources`| Refer to Kubernetes resource requests and limits for details. |


For details on the Kubernetes resource requests and limits, see the [Manage Compute Resources Container](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) document.
