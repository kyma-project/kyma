# Etcd Backup

## Overview

The etcd-backup trigger a backup process of etcd-cluster via etcd-backup-operator.

## Prerequisites

To set up the project, use these tools:
* Version 1.9 or higher of [Go](https://golang.org/dl/)
* The latest version of [Docker](https://www.docker.com/)
* The latest version of [Dep](https://github.com/golang/dep)

## Usage

This section explains how to use the Etcd Backup tool.

### Run a local version
To run the application without building the binary file, run this command:

```bash
#!/usr/bin/env bash

export APP_LOGGER_LEVEL="debug"
export APP_KUBECONFIG_PATH="/Users/{User}/.kube/config"

export APP_WORKING_NAMESPACE="kyma-system"
export APP_ABS_CONTAINER_NAME=<container_name>
export APP_ABS_SECRET_NAME=<secret_name>
export APP_BLOB_PREFIX=<prefix_name>

export APP_BACKUP_CONFIG_MAP_NAME_FOR_TRACING="sc-recorded-etcd-backup-data"
export APP_BACKUP_ETCD_ENDPOINTS="<endpoints>"

go run main.go
```

For the description of the available environment variables, see the **Use environment variables** section.

### Use environment variables
Use the following environment variables to configure the application:

| Name | Required | Default | Description |
|-----|---------|--------|------------|
| **APP_LOGGER_LEVEL** | No | `info` | Show detailed logs in the application. |
| **APP_KUBECONFIG_PATH** | No |  | The path to the `kubeconfig` file that you need to run an application outside of the cluster. |
| **APP_WORKING_NAMESPACE** | Yes |  | The name of the Namespace where Etcd Backup application is executed. |
| **APP_ABS_CONTAINER_NAME** | Yes |  | The Azure Blob Storage container name where the backup should be saved. |
| **APP_ABS_SECRET_NAME** | Yes |  | The name of the secret object that stores the Azure storage credential. |
| **APP_BLOB_PREFIX** | Yes |  | The name of the blob prefix which should be used for saved backup. Basically it should be the name of application for which backup is performed e.g. **service-catalog** |
| **APP_BACKUP_ETCD_ENDPOINTS** | Yes |  | The endpoints of an etcd cluster. When multiple endpoints are given, the backup operator retrieves the backup from the endpoint that has the most up-to-date state. The given endpoints must belong to the same etcd cluster. |
| **APP_BACKUP_CONFIG_MAP_NAME_FOR_TRACING** | Yes |  | The name of the ConfigMap where the path to the ABS backup is saved (only from the last success). |


## Development

Use the `before-commit.sh` script to test your changes before each commit.
