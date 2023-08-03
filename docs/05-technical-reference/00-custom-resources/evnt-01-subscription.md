---
title: Subscription
---

The `subscriptions.eventing.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to subscribe to events. To get the up-to-date CRD and show the output in the YAML format, run this command:

`kubectl get crd subscriptions.eventing.kyma-project.io -o yaml`

## Sample custom resource

This sample Subscription custom resource (CR) subscribes to an event called `order.created.v1`.

> **WARNING:** Prohibited characters in event names under the **spec.types** property, are not supported in some backends. If any are detected, Eventing will remove them. Read [Event names](../evnt-01-event-names.md#event-name-cleanup) for more information.

> **NOTE:** Both the subscriber and the Subscription should exist in the same Namespace.

```yaml
apiVersion: eventing.kyma-project.io/v1alpha2
kind: Subscription
metadata:
  name: test
  namespace: test
spec:
  typeMatching: standard
  source: commerce
  types:
    - order.created.v1
  sink: http://test.test.svc.cluster.local
  config:
    maxInFlightMessages: "10"
```

## Custom resource parameters

This table lists all the possible parameters of a given resource together with their descriptions:

<!-- TABLE-START -->
### Subscription.eventing.kyma-project.io/v1alpha2

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **config**  | map\[string\]string | Map of configuration options that will be applied on the backend. |
| **id**  | string | Unique identifier of the Subscription, read-only. |
| **sink** (required) | string | Kubernetes Service that should be used as a target for the events that match the Subscription. Must exist in the same Namespace as the Subscription. |
| **source** (required) | string | Defines the origin of the event. |
| **typeMatching**  | string | Defines how types should be handled.<br /> - `standard`: backend-specific logic will be applied to the configured source and types.<br /> - `exact`: no further processing will be applied to the configured source and types. |
| **types** (required) | \[\]string | List of event types that will be used for subscribing on the backend. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **backend**  | object | Backend-specific status which is applicable to the active backend only. |
| **backend.&#x200b;apiRuleName**  | string | Name of the APIRule which is used by the Subscription. |
| **backend.&#x200b;emsSubscriptionStatus**  | object | Status of the Subscription as reported by EventMesh. |
| **backend.&#x200b;emsSubscriptionStatus.&#x200b;lastFailedDelivery**  | string | Timestamp of the last failed delivery. |
| **backend.&#x200b;emsSubscriptionStatus.&#x200b;lastFailedDeliveryReason**  | string | Reason for the last failed delivery. |
| **backend.&#x200b;emsSubscriptionStatus.&#x200b;lastSuccessfulDelivery**  | string | Timestamp of the last successful delivery. |
| **backend.&#x200b;emsSubscriptionStatus.&#x200b;status**  | string | Status of the Subscription as reported by the backend. |
| **backend.&#x200b;emsSubscriptionStatus.&#x200b;statusReason**  | string | Reason for the current status. |
| **backend.&#x200b;emsTypes**  | \[\]object | List of mappings from event type to EventMesh compatible types. Used only with EventMesh as the backend. |
| **backend.&#x200b;emsTypes.&#x200b;eventMeshType** (required) | string | Event type that is used on the EventMesh backend. |
| **backend.&#x200b;emsTypes.&#x200b;originalType** (required) | string | Event type that was originally used to subscribe. |
| **backend.&#x200b;emshash**  | integer | Hash used to identify an EventMesh Subscription retrieved from the server without the WebhookAuth config. |
| **backend.&#x200b;ev2hash**  | integer | Checksum for the Subscription custom resource. |
| **backend.&#x200b;eventMeshLocalHash**  | integer | Hash used to identify an EventMesh Subscription posted to the server without the WebhookAuth config. |
| **backend.&#x200b;externalSink**  | string | Webhook URL used by EventMesh to trigger subscribers. |
| **backend.&#x200b;failedActivation**  | string | Provides the reason if a Subscription failed activation in EventMesh. |
| **backend.&#x200b;types**  | \[\]object | List of event type to consumer name mappings for the NATS backend. |
| **backend.&#x200b;types.&#x200b;consumerName**  | string | Name of the JetStream consumer created for the event type. |
| **backend.&#x200b;types.&#x200b;originalType** (required) | string | Event type that was originally used to subscribe. |
| **backend.&#x200b;webhookAuthHash**  | integer | Hash used to identify the WebhookAuth of an EventMesh Subscription existing on the server. |
| **conditions**  | \[\]object | Current state of the Subscription. |
| **conditions.&#x200b;lastTransitionTime**  | string | Defines the date of the last condition status change. |
| **conditions.&#x200b;message**  | string | Provides more details about the condition status change. |
| **conditions.&#x200b;reason**  | string | Defines the reason for the condition status change. |
| **conditions.&#x200b;status** (required) | string | Status of the condition. The value is either `True`, `False`, or `Unknown`. |
| **conditions.&#x200b;type**  | string | Short description of the condition. |
| **ready** (required) | boolean | Overall readiness of the Subscription. |
| **types** (required) | \[\]object | List of event types after cleanup for use with the configured backend. |
| **types.&#x200b;cleanType** (required) | string | Event type after it was cleaned up from backend compatible characters. |
| **types.&#x200b;originalType** (required) | string | Event type as specified in the Subscription spec. |

### Subscription.eventing.kyma-project.io/v1alpha1

>**CAUTION**: The v1alpha1 API version is deprecated as of Kyma 2.14.X.

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **config**  | object | Defines additional configuration for the active backend. |
| **config.&#x200b;maxInFlightMessages**  | integer | Defines how many not-ACKed messages can be in flight simultaneously. |
| **filter** (required) | object | Defines which events will be sent to the sink. |
| **filter.&#x200b;dialect**  | string | Contains a `URI-reference` to the CloudEvent filter dialect. See [here](https://github.com/cloudevents/spec/blob/main/subscriptions/spec.md#3241-filter-dialects) for more details. |
| **filter.&#x200b;filters** (required) | \[\]object | Defines the BEB filter element as a combination of two CE filter elements. |
| **filter.&#x200b;filters.&#x200b;eventSource** (required) | object | Defines the source of the CE filter. |
| **filter.&#x200b;filters.&#x200b;eventSource.&#x200b;property** (required) | string | Defines the property of the filter. |
| **filter.&#x200b;filters.&#x200b;eventSource.&#x200b;type**  | string | Defines the type of the filter. |
| **filter.&#x200b;filters.&#x200b;eventSource.&#x200b;value** (required) | string | Defines the value of the filter. |
| **filter.&#x200b;filters.&#x200b;eventType** (required) | object | Defines the type of the CE filter. |
| **filter.&#x200b;filters.&#x200b;eventType.&#x200b;property** (required) | string | Defines the property of the filter. |
| **filter.&#x200b;filters.&#x200b;eventType.&#x200b;type**  | string | Defines the type of the filter. |
| **filter.&#x200b;filters.&#x200b;eventType.&#x200b;value** (required) | string | Defines the value of the filter. |
| **id**  | string | Unique identifier of the Subscription, read-only. |
| **protocol**  | string | Defines the CE protocol specification implementation. |
| **protocolsettings**  | object | Defines the CE protocol settings specification implementation. |
| **protocolsettings.&#x200b;contentMode**  | string | Defines the content mode for eventing based on BEB. The value is either `BINARY`, or `STRUCTURED`. |
| **protocolsettings.&#x200b;exemptHandshake**  | boolean | Defines if the exempt handshake for eventing is based on BEB. |
| **protocolsettings.&#x200b;qos**  | string | Defines the quality of service for eventing based on BEB. |
| **protocolsettings.&#x200b;webhookAuth**  | object | Defines the Webhook called by an active subscription on BEB. |
| **protocolsettings.&#x200b;webhookAuth.&#x200b;clientId** (required) | string | Defines the clientID for OAuth2. |
| **protocolsettings.&#x200b;webhookAuth.&#x200b;clientSecret** (required) | string | Defines the Client Secret for OAuth2. |
| **protocolsettings.&#x200b;webhookAuth.&#x200b;grantType** (required) | string | Defines the grant type for OAuth2. |
| **protocolsettings.&#x200b;webhookAuth.&#x200b;scope**  | \[\]string | Defines the scope for OAuth2. |
| **protocolsettings.&#x200b;webhookAuth.&#x200b;tokenUrl** (required) | string | Defines the token URL for OAuth2. |
| **protocolsettings.&#x200b;webhookAuth.&#x200b;type**  | string | Defines the authentication type. |
| **sink** (required) | string | Kubernetes Service that should be used as a target for the events that match the Subscription. Must exist in the same Namespace as the Subscription. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **apiRuleName**  | string | Defines the name of the APIRule which is used by the Subscription. |
| **cleanEventTypes** (required) | \[\]string | CleanEventTypes defines the filter's event types after cleanup to use it with the configured backend. |
| **conditions**  | \[\]object | Current state of the Subscription. |
| **conditions.&#x200b;lastTransitionTime**  | string | Defines the date of the last condition status change. |
| **conditions.&#x200b;message**  | string | Provides more details about the condition status change. |
| **conditions.&#x200b;reason**  | string | Defines the reason for the condition status change. |
| **conditions.&#x200b;status** (required) | string | Status of the condition. The value is either `True`, `False`, or `Unknown`. |
| **conditions.&#x200b;type**  | string | Short description of the condition. |
| **config**  | object | Defines the configurations that have been applied to the eventing backend when creating this Subscription. |
| **config.&#x200b;maxInFlightMessages**  | integer | Defines how many not-ACKed messages can be in flight simultaneously. |
| **emsSubscriptionStatus**  | object | Defines the status of the Subscription in EventMesh. |
| **emsSubscriptionStatus.&#x200b;lastFailedDelivery**  | string | Timestamp of the last failed delivery. |
| **emsSubscriptionStatus.&#x200b;lastFailedDeliveryReason**  | string | Reason for the last failed delivery. |
| **emsSubscriptionStatus.&#x200b;lastSuccessfulDelivery**  | string | Timestamp of the last successful delivery. |
| **emsSubscriptionStatus.&#x200b;subscriptionStatus**  | string | Status of the Subscription as reported by EventMesh. |
| **emsSubscriptionStatus.&#x200b;subscriptionStatusReason**  | string | Reason for the current status. |
| **emshash**  | integer | Defines the checksum for the Subscription in EventMesh. |
| **ev2hash**  | integer | Defines the checksum for the Subscription custom resource. |
| **externalSink**  | string | Defines the webhook URL which is used by EventMesh to trigger subscribers. |
| **failedActivation**  | string | Defines the reason if a Subscription failed activation in EventMesh. |
| **ready** (required) | boolean | Overall readiness of the Subscription. |

<!-- TABLE-END -->

## Related resources and components

These components use this CR:

| Component   |   Description |
|-------------|---------------|
| [Eventing Controller](../00-architecture/evnt-01-architecture.md#eventing-controller) | The Eventing Controller reconciles on Subscriptions and creates a connection between subscribers and the Eventing backend. |
| [Event Publisher Proxy](../00-architecture/evnt-01-architecture.md#event-publisher-proxy) | The Event Publisher Proxy reads the Subscriptions to find out how events are used for each Application. |
