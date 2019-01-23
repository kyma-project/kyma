---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name    | Description      |
| ------------ | ---------------- |
| `all-in-one` | The default plan |

## Provision

Provisions a blob storage account and create a container within the account.

### Provisioning Parameters

| Parameter Name          | Type                | Description                                                  | Required | Default Value                                                |
| ----------------------- | ------------------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| **location**              | `string`            | The Azure region in which to provision applicable resources. | Yes      |                                                              |
| **resourceGroup**         | `string`            | The (new or existing) resource group with which to associate new resources. | Yes      |                                                              |
| **enableNonHttpsTraffic** | `string`            | Specify whether non-https traffic is enabled. Allowed values:["enabled", "disabled"]. | No       | If not provided, "disabled" will be used as the default value. That is, only https traffic is allowed. |
| **accessTier**            | `string`            | The access tier used for billing.    Allowed values: ["Hot", "Cool"]. Hot storage is optimized for storing data that is accessed frequently ,and cool storage is optimized for storing data that is infrequently accessed and stored for at least 30 days. | No       | If not provided, "Hot" will be used as the default value.    |
| **accountType**           | `string`            | A combination of account kind and   replication strategy. All possible values: ["Standard_LRS", "Standard_GRS", "Standard_RAGRS"]. | No       | If not provided, "Standard_LRS" will be used as the default value for all plans. |
| **containerName**         | `string`            | The name of the container which will be created inside the storage account. This name may only contain lowercase letters, numbers, and hyphens, and must begin with a letter or a number. Each hyphen must be preceded and followed by a non-hyphen character. The length of the name must between 3 and 63. | No       | If not provided, a random name will be generated as the container name. |
| **tags**                  | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No       | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |

## Update

Updates an existing storage account.

### Updating parameters

| Parameter Name            | Type                | Description                                                  | Required |
| ------------------------- | ------------------- | ------------------------------------------------------------ | -------- |
| **enableNonHttpsTraffic** | `string`            | Specify whether non-https traffic is enabled. Allowed values:["enabled", "disabled"]. | No       |
| **accessTier**              | `string`            | The access tier used for billing.    Allowed values: ["Hot", "Cool"]. Hot storage is optimized for storing data that is accessed frequently ,and cool storage is optimized for storing data that is infrequently accessed and stored for at least 30 days. | No        |
| **accountType**             | `string`            | A combination of account kind and   replication strategy.    | No       |
| **tags**                    | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No       |

## Credentials

Binding returns the following connection details and shared credentials:

| Field Name                   | Type     | Description                                           |
| ---------------------------- | -------- | ----------------------------------------------------- |
| **storageAccountName**         | `string` | The storage account name.                             |
| **accessKey**                  | `string` | A key (password) for accessing the storage account.   |
| **primaryBlobServiceEndPoint** | `string` | Primary blob service end point.                       |
| **containerName**              | `string` | The name of the container within the storage account. |
