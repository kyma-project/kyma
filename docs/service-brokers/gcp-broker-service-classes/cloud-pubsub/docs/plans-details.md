---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Beta Plan` | Pub/Sub plan for the Beta release of the Google Cloud Platform Service Broker |

## Provisioning parameters

Provisioning an instance creates a new pub/sub topic. These are the input parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **topicId** | `string` | A user-specified Pub/Sub topic ID. Must be 3-255 characters long, start with an alphanumeric character, and contain only the following characters: letters, numbers, dashes, periods, underscores, tildes, percents or plus signs. Cannot start with `goog`. | YES | - |

## Update parameters:

The update parameters are the same as the provisioning parameters.

## Binding parameters:

Binding grants the provided service account with access on the Pub/Sub topic. Optionally, a new service account can be created and given access to the Pub/Sub topic. These are the binding parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `createServiceAccount` | `boolean` | Creates a new service account for Pub/Sub binding. | NO | `false` |
| `roles` | `array` | The list of Cloud Pub/Sub roles for the binding. Affects the level of access granted to the service account. These are the possible values of this parameter: `roles/pubsub.publisher`, `roles/pubsub.subscriber`, `roles/pubsub.viewer`, `roles/pubsub.editor`, `roles/pubsub.admin`. The items in the roles array must be unique, which means that you can specify a given role only once. | YES | - |
| `serviceAccount` | `string` | The GCP service account to which access is granted. | YES | - |
| `subscription` | `object` | A subscription resource. | NO | - |
| `subscription.ackDeadlineSeconds` | `integer` | This value is the maximum time after a subscriber receives a message\nbefore the subscriber should acknowledge the message. After message\ndelivery but before the ack deadline expires and before the message is\nacknowledged, it is an outstanding message and will not be delivered\nagain during that time (on a best-effort basis).\n\nFor pull subscriptions, this value is used as the initial value for the ack\ndeadline. To override this value for a given message, call\n`ModifyAckDeadline` with the corresponding `ack_id` if using\nnon-streaming pull or send the `ack_id` in a\n`StreamingModifyAckDeadlineRequest` if using streaming pull.\nThe minimum custom deadline you can specify is 10 seconds.\nThe maximum custom deadline you can specify is 600 seconds (10 minutes).\nIf this parameter is 0, a default value of 10 seconds is used.\n\nFor push delivery, this value is also used to set the request timeout for\nthe call to the push endpoint.\n\nIf the subscriber never acknowledges the message, the Pub/Sub\nsystem will eventually redeliver the message." minimum 0 maximum 600 | NO | - |
| `subscription.pushConfig` | `string` | A URL locating the endpoint to which messages should be pushed. If push delivery is used with this subscription, this field is\nused to configure it. An empty `pushConfig` signifies that the subscriber\nwill pull and ack messages using API methods. | NO | - |
| `subscription.pushConfig.attributes` | `object` | Endpoint configuration attributes.\n\nEvery endpoint has a set of API supported attributes that can be used to\ncontrol different aspects of the message delivery.\n\nThe currently supported attribute is `x-goog-version`, which you can\nuse to change the format of the pushed message. This attribute\nindicates the version of the data expected by the endpoint. This\ncontrols the shape of the pushed message (i.e., its fields and metadata).\nThe endpoint version is based on the version of the Pub/Sub API.\n\nIf not present during the `CreateSubscription` call, it will default to\nthe version of the API used to make such call. If not present during a\n`ModifyPushConfig` call, its value will not be changed. `GetSubscription`\ncalls will always return a valid version, even if the subscription was\ncreated without this attribute.\n\nThe possible values for this attribute are:\n\n* `v1beta1`: uses the push format defined in the v1beta1 Pub/Sub API.\n* `v1` or `v1beta2`: uses the push format defined in the v1 Pub/Sub API. | NO | - |
| `subscription.pushConfig.pushEndpoint` | `string` | A URL locating the endpoint to which messages should be pushed. | NO | - |
| `subscription.subscriptionId` | `string` | A user-specified Pubsub subscription ID. Must be 3-255 characters, start with an alphanumeric character, and contain only the following characters: letters, numbers, dashes, periods, underscores, tildes, percents or plus signs. | NO | - |
