# End-to-end Backup and Restore Tests

## Overview

This project contains end-to-end backup and restore tests for the Kyma installation. The tests run weekly on Prow to validate if the backup and restore process works for all components.
 

## Prerequisites

To set up the project, use these tools:

* Version 1.11 or higher of [Go](https://golang.org/dl/)
* Version 0.5 or higher of [Dep](https://github.com/golang/dep)
* The latest version of [Docker](https://www.docker.com/)
* [Ark](../../../resources/ark/README.md#details)

>**NOTE:** Use [these](../../../docs/backup/docs/03-01-backup-configuration.md) guidelines to configure Ark for a specific storage provider.


## Usage

The backup and restore [continuous integration flow](https://github.com/kyma-project/test-infra/blob/master/prow/scripts/cluster-integration/kyma-gke-end-to-end-test.sh) looks as follows:

1. Create a new Namespace.
2. Create new resources in the Namespace. The resources can be Namespace-scoped or cluster-wide.
3. Verify that the resources work.
4. Back up the Namespace and all the resources in it.
5. Remove the Namespace and all its resources.
6. Restore the Namespace and its resources from the backup.
7. Verify if the resources contained in the restored Namespace work.

### Use environment variables

Use the following environment variables to configure the tests:

| Name | Required | Default | Description |
|-----|:---------:|:--------:|------------|
| **DOMAIN** | NO | `kyma.local` | The domain where Kyma runs. |
| **USER_EMAIL** | YES | - | The email address for authentication in Dex. |
| **USER_PASSWORD** | YES | - | The password for authentication in Dex. |
| **KUBECONFIG** | NO | - | The path to the `kubeconfig` file needed to run tests outside the cluster. |
| **ALL_BACKUP_CONFIGURATION_FILE** | NO | `/all-backup.yaml` | The path to the `all-backup` configuration file. |
| **SYSTEM_BACKUP_CONFIGURATION_FILE** | NO | `/system-backup.yaml` | The path to the `system-backup` configuration file. |


## Development

This section presents how to add and run a new test. 

### Add a new test

Add a new test under the `backupe2e/{domain-name}` directory and implement the following interface:

```go
type BackupTest interface {
    CreateResources(namespace string)
    TestResources(namespace string)
    DeleteResources(namespace string)
}
```
The functions work as follows:

- The `TestResources` function validates if the test data works as expected. 
- The `CreateResources` function installs the required test data before the backup process starts.
- The `DeleteResources` function deletes the resources that are a part of the cluster before executing the test. The resources need to be deleted to test the restore process.

After the pipeline executes the backup and restore processes on the cluster, the `TestResources` function validates if the restore worked as expected.

The test creates a new Namespace called `{TestName}-{UUID}`. This Namespace should contain all resources created during the test. If required, the resources can be created in other Namespaces as well.

### Run end-to-end tests locally

> **NOTE:** Before running the test, configure Ark using [these](../../../docs/backup/docs/03-01-backup-configuration.md) guidelines.

Run tests:
```bash
env KUBECONFIG={KUBECONFIG} ALL_BACKUP_CONFIGURATION_FILE={ALL_BACKUP_PATH} SYSTEM_BACKUP_CONFIGURATION_FILE={SYSTEM_BACKUP_PATH} go test ./... -count=1 -timeout=0
```
where:
* `{KUBECONFIG}` is the path to the `kubeconfig` file.
* `{ALL_BACKUP_PATH}` is the path to the `all-backup.yaml` file.
* `{SYSTEM_BACKUP_PATH}` is the path to the `system-backup.yaml` file.

### Run tests using a Helm chart

Run the tests using Helm:

1. Prepare the data:

```bash
helm install deploy/chart/backup-test/ --name "backup-test" --namespace end-to-end --set global.ingress.domainName="$CLUSTER_DOMAIN" --set-file global.adminEmail=<(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.email}" | base64 --decode) --set-file global.adminPassword=<(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode)
```
2. Run tests:

```bash
helm test backup-test --timeout=0
```

### Run tests using Telepresence
[Telepresence](https://www.telepresence.io/) allows you to run tests locally while connecting a service to a remote Kubernetes cluster. It is helpful when the test needs access to other services in a cluster.

1. [Install Telepresence](https://www.telepresence.io/reference/install).
2. Run tests:
```bash
env KUBECONFIG={KUBECONFIG} ALL_BACKUP_CONFIGURATION_FILE={ALL_BACKUP_PATH} SYSTEM_BACKUP_CONFIGURATION_FILE={SYSTEM_BACKUP_PATH} telepresence --run go test ./... -count=1 -timeout=0
```
where:
* `{KUBECONFIG}` is the path to the `kubeconfig` file.
* `{ALL_BACKUP_PATH}` is the path to the `all-backup.yaml` file.
* `{SYSTEM_BACKUP_PATH}` is the path to the `system-backup.yaml` file.

### Verify the code

Use the `before-commit.sh` script or the `make build` command to test your changes before each commit.

### Project structure

The repository has the following structure:

```
├── deploy                          # The Helm chart for deploying the backup test application with configuration
├── backupe2e                       # The package where backup tests are defined. Put your test here.
├── utils                           # The directory which contains all secondary Go packages.
├── restore_cluster_backup_test.go  # The entrypoint for backup test runner
├── Gopkg.toml                      # A dep manifest
└── Gopkg.lock                      # A dep lock which is generated automatically. Do not edit it.

```
