# Etcd Backup

## Overview

The Etcd Backup triggers a backup process of the `etcd` cluster using the etcd-backup-operator. This application also removes the old backup files from the Azure Blob Storage (ABS). For more information, see the [Use environment variables](#use-environment-variables) section.

## Prerequisites

To set up the project, download these tools:

* [Go](https://golang.org/dl/) 1.11.4
* [Dep](https://github.com/golang/dep) v0.5.0
* [Docker](https://www.docker.com/)

These Go and Dep versions are compliant with the `buildpack` used by Prow. For more details read [this](https://github.com/kyma-project/test-infra/blob/master/prow/images/buildpack-golang/README.md) document.

## Usage

This section explains how to use the Etcd Backup tool.

### Run a local version
To run the application without building the binary file, run this command:

```bash
#!/usr/bin/env bash

export APP_LOGGER_LEVEL="debug"
export APP_KUBECONFIG_PATH="/Users/{User}/.kube/config"

export APP_WORKING_NAMESPACE="kyma-system"
export APP_ABS_CONTAINER_NAME={container_name}
export APP_ABS_SECRET_NAME={secret_name}
export APP_BLOB_PREFIX={prefix_name}

export APP_BACKUP_CONFIG_MAP_NAME_FOR_TRACING="sc-recorded-etcd-backup-data"
export APP_BACKUP_ETCD_ENDPOINTS="{endpoints}"
export APP_BACKUP_CLIENT_TLS_SECRET="core-service-catalog-etcd-etcd-client-tls" # If the TLS for the etcd is enabled.

export APP_CLEANER_LEAVE_MIN_NEWEST_BACKUP_BLOBS="3"
export APP_CLEANER_EXPIRATION_BLOB_TIME="24h"

go run main.go
```

For the description of the available environment variables, see the **Use environment variables** section.

### Use environment variables
Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|-----|---------|--------|------------|
| **APP_LOGGER_LEVEL** | No | `info` | Show detailed logs in the application. |
| **APP_KUBECONFIG_PATH** | No |  | The path to the `kubeconfig` file that you need to run an application outside of the cluster. |
| **APP_WORKING_NAMESPACE** | Yes |  | The Namespace where the Etcd Backup application is executed. |
| **APP_ABS_CONTAINER_NAME** | Yes |  | The Azure Blob Storage container to store the backup. |
| **APP_ABS_SECRET_NAME** | Yes |  | The name of the Secret object that stores the Azure storage credential. |
| **APP_BLOB_PREFIX** | Yes |  | The name of the blob prefix to use to save the backup. Basically, it should be the name of the application for which the system performs the backup e.g. **service-catalog** |
| **APP_BACKUP_ETCD_ENDPOINTS** | Yes |  | The endpoints of the `etcd` cluster. When there are multiple endpoints, the backup operator retrieves the backup from the endpoint that has the most up-to-date state. The given endpoints must belong to the same etcd cluster. Multiple endpoints should be separated by comma. |
| **APP_BACKUP_CONFIG_MAP_NAME_FOR_TRACING** | Yes |  | The name of the ConfigMap where the path to the last successful ABS backup is saved. |
| **APP_CLEANER_LEAVE_MIN_NEWEST_BACKUP_BLOBS** | Yes |  | The number of blobs which should not be deleted even if they are treated as expired. |
| **APP_CLEANER_EXPIRATION_BLOB_TIME** | Yes |  | The duration used to check if a given blob should be deleted. If the **blob.LastModified** is earlier than the current time reduced by the **APP_CLEANER_EXPIRATION_BLOB_TIME** then the blob is removed. |

## Development

Use the `before-commit.sh` script to test your changes before each commit.
