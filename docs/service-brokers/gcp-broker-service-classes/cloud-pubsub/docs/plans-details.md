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

Provisioning an instance creates a new Pub/Sub topic. These are the input parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **topicId** | `string` | A user-specified Pub/Sub topic ID. Must be 3-255 characters long, start with an alphanumeric character, and contain only the following characters: letters, numbers, dashes, periods, underscores, tildes, percents or plus signs. Cannot start with `goog`. | YES | - |

## Update parameters

The update parameters are the same as the provisioning parameters.

## Binding parameters

Binding grants the provided service account access to the Pub/Sub topic. Optionally, you can create a new service account and add the access to the Pub/Sub topic. These are the binding parameters:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **createServiceAccount** | `boolean` | Creates a new service account for Pub/Sub binding. | NO | `false` |
| **roles** | `array` | The list of Cloud Pub/Sub roles for the binding. Affects the level of access granted to the service account. These are the possible values of this parameter: `roles/pubsub.publisher`, `roles/pubsub.subscriber`, `roles/pubsub.viewer`, `roles/pubsub.editor`, `roles/pubsub.admin`. The items in the roles array must be unique, which means that you can specify a given role only once. | YES | - |
| **serviceAccount** | `string` | The GCP service account to which access is granted. | YES | - |
| **subscription** | `object` | A subscription resource. For more information, go to the **Subscription properties** section. | NO | - |

### Subscription properties

These are the **Subscription** properties:

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| **ackDeadlineSeconds** | `integer` | This value is the maximum time after a subscriber receives or acknowledges the message. After that time, or before the acknowledgement, the message is outstanding and is not delivered. For pull subscriptions, this value is used as the initial value for the **ackDeadline**. To override this value for a given message, call **ModifyAckDeadline** with the corresponding **ack_id**, if you use a non-streaming pull, or send the **ack_id** in a **StreamingModifyAckDeadlineRequest** if you use a streaming pull. The minimum custom deadline you can specify is `10` seconds and the maximum is `600` seconds. For push delivery, this value is also used to set the request timeout for the call to the push endpoint. If the subscriber never acknowledges the message, the Pub/Sub system eventually redelivers the message. | NO | `10` |
| **pushConfig** | `string` | A URL locating the endpoint to which messages are pushed. If push delivery is used with this subscription, this field is used to configure it. An empty **pushConfig** signifies that the subscriber pulls and acknowledges messages using API methods. | NO | - |
| **pushConfig.attributes** | `object` | Endpoint configuration attributes. Every endpoint has a set of API supported attributes that you can use to control different aspects of the message delivery. The currently supported attribute is **x-goog-version**, which you can use to change the format of the pushed message. This attribute indicates the version of the data expected by the endpoint. This controls the shape of the pushed message, such as its fields and metadata. The endpoint version is based on the version of the Pub/Sub API. If not present during the **CreateSubscription** call, it defaults to the version of the API used to make such call. If not present during a **ModifyPushConfig** call, its value will not be changed. **GetSubscription** calls always return a valid version, even if the subscription was created without this attribute. The possible values for this attribute are `v1beta1`, which uses the push format defined in the v1beta1 Pub/Sub API, or `v1beta2`, which uses the push format defined in the v1 Pub/Sub API. | NO | - |
| **pushConfig.pushEndpoint** | `string` | A URL locating the endpoint to which messages are pushed. | NO | - |
| **subscriptionId** | `string` | A user-specified Pub/Sub subscription ID. Must be 3-255 characters, start with an alphanumeric character, and contain only the following characters: letters, numbers, dashes, periods, underscores, tildes, percents or plus signs. | NO | - |

### Credentials

Binding returns the following connection details and credentials:

| Parameter Name | Type | Description |
|----------------|------|-------------|
| **privateKeyData** | `JSON Object` | The service account OAuth information. |
| **projectId** | `string` | The ID of the project. |
| **serviceAccount** | `string` | The GCP service account to which access is granted. |
| **subscriptionId** | `string` | The ID of the subscription. |
| **topicId** | `string` | The ID of the topic. |
