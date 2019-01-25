---
title: Services and Plans
type: Details
---

## Service description

The `azure-iot-hub` service consist of the following plan:

| Plan Name     | Description                                                  |
| ------------- | ------------------------------------------------------------ |
| `free`        | IoT hub Free Tier - max 8,000 messages per day.              |
| `basic-b1`    | IoT hub Basic B1 Tier - max 400,000 messages per unit per day. |
| `basic-b2`    | IoT hub Basic B2 Tier - max 6,000,000 messages per unit per day. |
| `basic-b3`    | IoT hub Basic B3 Tier - max 300,000,000 messages per unit per day. |
| `standard-s1` | IoT hub Standard S1 Tier - max 400,000 messages per unit per day. |
| `standard-s2` | IoT hub Standard S2 Tier - max 6,000,000 messages per unit per day. |
| `standard-s3` | IoT hub Standard S3 Tier - max 300,000,000 messages per unit per day. |

## Provision

Provisions a new IoT Hub.

### Provisioning Parameters

| Parameter Name   | Type                | Description                                                  | Required | Default Value                                                |
| ---------------- | ------------------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| **location**       | `string`            | The Azure region in which to provision applicable resources. | Y        |                                                              |
| **resourceGroup**  | `string`            | The (new or existing) resource group with which to associate new resources. | Y        |                                                              |
| **tags**           | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | N        | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |
| **units**          | `int`               | Number of IoT hub units. Each IoT Hub is provisioned with a certain number of units in a specific tier. The tier and number of units determine the maximum daily quota of messages that you can send. **Note**: for plan `free`, this field is invalid; for plan `standard-s3`, allowed values are from 1 to 10; for other plans, allowed values are from 1 to 200. | N        | If not provided, `1 `will be used as default value.          |
| **partitionCount** | `int`               | The number of partitions relates the device-to-cloud messages to the number of simultaneous readers of these messages. Most IoT hubs only need four partitions. **Note**: for plan `free`, this field is invalid; for plan `basic-*`, allowed values are from 2 to 8; for plan `standard-*`, allowed values are from 2 to 32. | N        | If not provided, `4` will be used as default value. For plan `free`, this field cannot be provided and `2` will be used. |

## Bind

Returns a copy of one shared set of credentials.

### Binding Parameters

This binding operation does not support any parameters.

### Credentials

Binding returns the following connection details and shared credentials:

| Field Name         | Type     | Description                                                  |
| ------------------ | -------- | ------------------------------------------------------------ |
| **iotHubName**       | `string` | The name of the IoT Hub.                                     |
| **hostName**         | `string` | Hostname.                                                    |
| **keyName**          | `string` | Name of the key. Currently, it will always be `iothubowner`  |
| **key**              | `string` | Key of the IoT Hub. Currently, it will always be the primary key of `iothubowner` and having "Registry Write, Service Connect, Device Connect" rights. |
| **connectionString** | `string` | The connection string.                                       |

