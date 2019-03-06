---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description                     |
| --------- | ------------------------------- |
| `topic`   | New topic in existing namespace |

## Provision

Provisions a new topic in an existing namespace. 

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name      | Type     | Description                                                  | Required | Default Value                                                |
| ------------------- | -------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| **parentAlias**       | `string` | Specifies the alias of the namespace in which the  topic should be provisioned. **Note**: the parent must be a service-bus-namespace instance with `standard` or `premium` plan. | Yes      |                                                              |
| **topicName**         | `string` | The name of the topic                                        | No        | If not provided, a random name will be generated as the topic name. |
| **maxTopicSize**      | `int`    | The maximum size of the topic in megabytes, which is the size of memory allocated for the topic. | No       | 1024                                                         |
| **messageTimeToLive** | `string` | ISO 8601 default message timespan to live value. This is the duration after which the message expires, starting from when the message is sent to Service Bus. This is the default value used when TimeToLive is not set on a message itself. For example, `PT276H13M14S` sets the message to expire in 11 day 12 hour 13 minute 14 seconds. | No       | "PT336H"                                                     |

### Bind

Depending on the binding parameter, the binding operation will create a subscription in the topic or directly return the credential of the topic.

#### Binding Parameters

| Parameter Name       | Type     | Description                                                  | Required | Default Value |
| -------------------- | -------- | ------------------------------------------------------------ | -------- | ------------- |
| **subscriptionNeeded** | `string` | Specifies whether to create a subscription in the topic. Valid values are ["yes", "no"]. If set to "yes", a subscription having random name will be created in the topic; otherwise, it leaves everything unchanged. You may set this field to "yes" for message consumer, and set this field to "no" for message producer. | No       | "yes"         |

## Unbind

If **subscriptionNeeded** is set to "yes", deletes the created subscription; otherwise, does nothing.

### Credentials

Binding returns the following connection details and shared credentials:

| Field Name         | Type     | Description                                                  |
| ------------------ | -------- | ------------------------------------------------------------ |
| **connectionString** | `string` | Connection string.                                           |
| **primaryKey**       | `string` | Secret key (password).                                       |
| **namespaceName**    | `string` | The name of the namespace.                                   |
| **topicName**        | `string` | The name of the topic.                                       |
| **subscriptionName** | `string` | The name of the created subscription. Only appears when `subscriptionNeeded` is set to "yes". |

