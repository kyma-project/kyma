---
title: Subscription
type: Custom Resource
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
| **config**  | map\[string\]string | Config defines the configurations that can be applied to the eventing backend. |
| **id**  | string | ID is the unique identifier of Subscription, read-only. |
| **sink** (required) | string | The service that should be used as a target for the events that match the subscription. |
| **source** (required) | string | Source defines the origin of the event. |
| **typeMatching**  | string | Configurable TypeMatching defines how types should be handled.<br /> `standard`: backend specific logic will be applied to the configured source and types.<br /> `exact`: no further processing will be applied to the configured source and types. |
| **types** (required) | \[\]string | A list of event names. These names will be used for subscribing on the backend. |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **backend**  | object | Backend specific status which are only applicable to the active backend. |
| **backend.apiRuleName**  | string | APIRuleName defines the name of the APIRule which is used by the Subscription. |
| **backend.emsSubscriptionStatus**  | object | EmsSubscriptionStatus defines the status of Subscription in EventMesh. |
| **backend.emsSubscriptionStatus.lastFailedDelivery**  | string | LastFailedDelivery defines the timestamp of the last failed delivery. |
| **backend.emsSubscriptionStatus.lastFailedDeliveryReason**  | string | LastFailedDeliveryReason defines the reason of failed delivery. |
| **backend.emsSubscriptionStatus.lastSuccessfulDelivery**  | string | LastSuccessfulDelivery defines the timestamp of the last successful delivery. |
| **backend.emsSubscriptionStatus.status**  | string | Status defines the status of the Subscription. |
| **backend.emsSubscriptionStatus.statusReason**  | string | StatusReason defines the reason of the status. |
| **backend.emsTypes**  | \[\]object | EmsTypes is a list of mappings between event type and EventMesh compatible types. Only used with EventMesh as backend |
| **backend.emsTypes.eventMeshType** (required) | string | EventMeshType is the event type that is used on the event mesh backend |
| **backend.emsTypes.originalType** (required) | string | OriginalType is the event type originally used to subscribe |
| **backend.emshash**  | integer | Emshash defines the hash for the Subscription in EventType. |
| **backend.ev2hash**  | integer | Ev2hash defines the hash for the Subscription custom resource. |
| **backend.externalSink**  | string | ExternalSink defines the webhook URL which is used by EventMesh to trigger subscribers. |
| **backend.failedActivation**  | string | FailedActivation defines the reason if a Subscription had failed activation in EventMesh. |
| **backend.types**  | \[\]object | Types is a list of event type to consumer name mappings for the Nats backend |
| **backend.types.consumerName**  | string | ConsumerName is the name of the Jetstream consumer |
| **backend.types.originalType** (required) | string | OriginalType is the event type originally used to subscribe |
| **conditions**  | \[\]object | Current state of the Subscription |
| **conditions.lastTransitionTime**  | string |  |
| **conditions.message**  | string |  |
| **conditions.reason**  | string |  |
| **conditions.status** (required) | string | Status of the condition, one of True, False, Unknown |
| **conditions.type**  | string |  |
| **ready** (required) | boolean | Overall readiness of the Subscription |
| **types** (required) | \[\]object | Types defines the filter's event types after cleanup for use with the configured backend. |
| **types.cleanType** (required) | string | CleanType is the event type after it was cleaned up from backend compatible characters |
| **types.originalType** (required) | string | OriginalType is the event type specified in the subscription spec |

### Subscription.eventing.kyma-project.io/v1alpha1

**Spec:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **config**  | object | Config defines the configurations that can be applied to the eventing backend when creating this subscription |
| **config.maxInFlightMessages**  | integer |  |
| **filter** (required) | object | Filter defines the list of filters |
| **filter.dialect**  | string |  |
| **filter.filters** (required) | \[\]object | BEBFilter defines the BEB filter element as a combination of two CE filter elements. |
| **filter.filters.eventSource** (required) | object | EventSource defines the source of CE filter |
| **filter.filters.eventSource.property** (required) | string | Property defines the property of the filter |
| **filter.filters.eventSource.type**  | string | Type defines the type of the filter |
| **filter.filters.eventSource.value** (required) | string | Value defines the value of the filter |
| **filter.filters.eventType** (required) | object | EventType defines the type of CE filter |
| **filter.filters.eventType.property** (required) | string | Property defines the property of the filter |
| **filter.filters.eventType.type**  | string | Type defines the type of the filter |
| **filter.filters.eventType.value** (required) | string | Value defines the value of the filter |
| **id**  | string | ID is the unique identifier of Subscription, read-only. |
| **protocol**  | string | Protocol defines the CE protocol specification implementation |
| **protocolsettings**  | object | ProtocolSettings defines the CE protocol setting specification implementation |
| **protocolsettings.contentMode**  | string | ContentMode defines content mode for eventing based on BEB. |
| **protocolsettings.exemptHandshake**  | boolean | ExemptHandshake defines whether exempt handshake for eventing based on BEB. |
| **protocolsettings.qos**  | string | Qos defines quality of service for eventing based on BEB. |
| **protocolsettings.webhookAuth**  | object | WebhookAuth defines the Webhook called by an active subscription in BEB. |
| **protocolsettings.webhookAuth.clientId** (required) | string | ClientID defines clientID for OAuth2 |
| **protocolsettings.webhookAuth.clientSecret** (required) | string | ClientSecret defines client secret for OAuth2 |
| **protocolsettings.webhookAuth.grantType** (required) | string | GrantType defines grant type for OAuth2 |
| **protocolsettings.webhookAuth.scope**  | \[\]string | Scope defines scope for OAuth2 |
| **protocolsettings.webhookAuth.tokenUrl** (required) | string | TokenURL defines token URL for OAuth2 |
| **protocolsettings.webhookAuth.type**  | string | Type defines type of authentication |
| **sink** (required) | string | Sink defines endpoint of the subscriber |

**Status:**

| Parameter | Type | Description |
| ---- | ----------- | ---- |
| **apiRuleName**  | string | APIRuleName defines the name of the APIRule which is used by the Subscription |
| **cleanEventTypes** (required) | \[\]string | CleanEventTypes defines the filter's event types after cleanup for use with the configured backend |
| **conditions**  | \[\]object | Conditions defines the status conditions |
| **conditions.lastTransitionTime**  | string |  |
| **conditions.message**  | string |  |
| **conditions.reason**  | string |  |
| **conditions.status** (required) | string |  |
| **conditions.type**  | string |  |
| **config**  | object | Config defines the configurations that have been applied to the eventing backend when creating this subscription |
| **config.maxInFlightMessages**  | integer |  |
| **emsSubscriptionStatus**  | object | EmsSubscriptionStatus defines the status of Subscription in BEB |
| **emsSubscriptionStatus.lastFailedDelivery**  | string | LastFailedDelivery defines the timestamp of the last failed delivery |
| **emsSubscriptionStatus.lastFailedDeliveryReason**  | string | LastFailedDeliveryReason defines the reason of failed delivery |
| **emsSubscriptionStatus.lastSuccessfulDelivery**  | string | LastSuccessfulDelivery defines the timestamp of the last successful delivery |
| **emsSubscriptionStatus.subscriptionStatus**  | string | SubscriptionStatus defines the status of the Subscription |
| **emsSubscriptionStatus.subscriptionStatusReason**  | string | SubscriptionStatusReason defines the reason of the status |
| **emshash**  | integer | Emshash defines the hash for the Subscription in BEB |
| **ev2hash**  | integer | Ev2hash defines the hash for the Subscription custom resource |
| **externalSink**  | string | ExternalSink defines the webhook URL which is used by BEB to trigger subscribers |
| **failedActivation**  | string | FailedActivation defines the reason if a Subscription had failed activation in BEB |
| **ready** (required) | boolean | Ready defines the overall readiness status of a subscription |

<!-- TABLE-END -->

## Related resources and components

These components use this CR:

| Component   |   Description |
|-------------|---------------|
| [Eventing Controller](../00-architecture/evnt-01-architecture.md#eventing-controller) | The Eventing Controller reconciles on Subscriptions and creates a connection between subscribers and the Eventing backend. |
| [Event Publisher Proxy](../00-architecture/evnt-01-architecture.md#event-publisher-proxy) | The Event Publisher Proxy reads the Subscriptions to find out how events are used for each Application. |
