---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description                     |
| --------- | ------------------------------- |
| `queue`   | New queue in existing namespace |

## Provision

Provisions a new queue in an existing namespace. 

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name      | Type     | Description                                                  | Required | Default Value                                                |
| ------------------- | -------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| **parentAlias**       | `string` | Specifies the alias of the namespace in which the  queue should be provisioned. | Yes      |                                                              |
| **queueName**         | `string` | The name of the queue.                                       | No       | If not provided, a random name will be generated as the queue name. |
| **maxQueueSize**      | `int`    | The maximum size of the queue in megabytes, which is the size of memory allocated for the queue. | No       | 1024                                                         |
| **messageTimeToLive** | `string` | ISO 8601 default message timespan to live value. This is the duration after which the message expires, starting from when the message is sent to Service Bus. This is the default value used when TimeToLive is not set on a message itself. For example, `PT276H13M14S` sets the message to expire in 11 day 12 hour 13 minute 14 seconds. | No       | "PT336H"                                                     |
| **lockDuration**      | `string` | ISO 8601 timespan duration of a peek-lock; that is, the amount of time that the message is locked for other receivers. The lock duration time window can range from 5 seconds to 5 minutes. For example, `PT2M30S` sets the lock duration time to 2 minutes 30 seconds. | No       | "PT30S"                                                      |

### Credentials

Binding returns the following connection details and shared credentials:

| Field Name         | Type     | Description                |
| ------------------ | -------- | -------------------------- |
| **connectionString** | `string` | Connection string.         |
| **primaryKey**       | `string` | Secret key (password).     |
| **namespaceName**    | `string` | The name of the namespace. |
| **queueName**        | `string` | The name of the queue.     |
