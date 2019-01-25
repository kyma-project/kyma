---
title: Services and Plans
type: Details
---

## Service description

The `azure-storage-general-purpose-v2-storage-account` service provides the following plan names and descriptions:

| Plan Name | Description                                       |
| --------- | ------------------------------------------------- |
| `account` | This plan provisions a general purpose v2 account.|

## Provision

Provisions a general purpose v2 storage account.

### Provisioning Parameters

| Parameter Name          | Type                | Description                                                  | Required | Default Value                                                |
| ----------------------- | ------------------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| **location**              | `string`            | The Azure region in which to provision applicable resources. | Yes        |                                                              |
| **resourceGroup**         | `string`            | The (new or existing) resource group with which to associate new resources. | Yes        |                                                              |
| **enableNonHttpsTraffic** | `string`            | Specify whether non-https traffic is enabled. Allowed values:["enabled", "disabled"]. | No        | If not provided, "disabled" will be used as the default value. That is, only https traffic is allowed. |
| **accessTier**            | `string`            | The access tier used for billing.    Allowed values: ["Hot", "Cool"]. Hot storage is optimized for storing data that is accessed frequently ,and cool storage is optimized for storing data that is infrequently accessed and stored for at least 30 days. **Note** : `accountType` "Premium_LRS" only supports "Hot" in this field | No        | If not provided, "Hot" will be used as the default value.    |
| **accountType**           | `string`            | A combination of account kind and   replication strategy. All possible values: ["Standard_LRS", "Standard_GRS", "Standard_RAGRS", "Standard_ZRS", "Premium_LRS"]. **Note**: ZRS is only available in several regions, check [here](https://docs.microsoft.com/en-us/azure/storage/common/storage-redundancy-zrs#support-coverage-and-regional-availability) for allowed regions to use ZRS. | No        | If not provided, "Standard_LRS" will be used as the default value for all plans. |
| **tags**                  | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No        | Tags (even if none are specified) are automatically supplemented with `heritage: open-service-broker-azure`. |

## Update

Updates an existing storage account.

### Updating parameters

| Parameter Name            | Type                | Description                                                  | Required |
| ------------------------- | ------------------- | ------------------------------------------------------------ | -------- |
| **enableNonHttpsTraffic** | `string`            | Specify whether non-https traffic is enabled. Allowed values:["enabled", "disabled"]. | No        |
| **accessTier**              | `string`            | The access tier used for billing.    Allowed values: ["Hot", "Cool"]. Hot storage is optimized for storing data that is accessed frequently ,and cool storage is optimized for storing data that is infrequently accessed and stored for at least 30 days. **Note** : `accountType` "Premium_LRS" only supports "Hot" in this field. | No       |
| **accountType**             | `string`            | A combination of account kind and   replication strategy. You can only update ["Standard_LRS", "Standard_GRS", "Standard_RAGRS"] accounts to one of ["Standard_LRS", "Standard_GRS", "Standard_RAGRS"]. For "Standard_ZRS" and "Premium_LRS" accounts, they are not updatable. | No       |
| **tags**                    | `map[string]string` | Tags to be applied to new resources, specified as key/value pairs. | No       |

## Credentials

Binding returns the following connection details and shared credentials:

| Field Name                    | Type     | Description                                         |
| ----------------------------- | -------- | --------------------------------------------------- |
| **storageAccountName**          | `string` | The storage account name.                           |
| **accessKey**                   | `string` | A key (password) for accessing the storage account. |
| **primaryBlobServiceEndPoint**  | `string` | Primary blob service end point.                     |
| **primaryTableServiceEndPoint** | `string` | Primary table service end point.                    |
| **primaryFileServiceEndPoint**  | `string` | Primary file service end point.                     |
| **primaryQueueServiceEndPoint** | `string` | Primary queue service end point. 
