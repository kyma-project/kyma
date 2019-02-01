---
title: Services and Plans
type: Details
---

## Service description

The service provides the following plan names and descriptions:

| Plan Name | Description      |
| --------- | ---------------- |
| `container` | This plan creates a container inside an existing blob storage account. |

## Provision

Create a blob container inside an blob storage account.

### Provisioning parameters

These are the provisioning parameters:

| Parameter Name  | Type     | Description                                                  | Required | Default Value                                                |
| --------------- | -------- | ------------------------------------------------------------ | -------- | ------------------------------------------------------------ |
| **parentAlias**   | `string` | Specifies the alias of the blob storage account upon which the  should be provisioned. | Y        |                                                              |
| **containerName** | `string` | The name of the container which will be created inside the storage account. This name may only contain lowercase letters, numbers, and hyphens, and must begin with a letter or a number. Each hyphen must be preceded and followed by a non-hyphen character. The length of the name must between 3 and 63. | N        | If not provided, a random name will be generated as the container name. |

### Credentials

Binding returns the following connection details and shared credentials:

| Field Name                   | Type     | Description                                           |
| ---------------------------- | -------- | ----------------------------------------------------- |
| **storageAccountName**         | `string` | The storage account name.                             |
| **accessKey**                  | `string` | A key (password) for accessing the storage account.   |
| **primaryBlobServiceEndPoint** | `string` | Primary blob service end point.                       |
| **containerName**              | `string` | The name of the container within the storage account. |
