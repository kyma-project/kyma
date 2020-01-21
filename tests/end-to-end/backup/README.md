# End-to-end Backup Tests

## Overview

This project contains end-to-end backup tests for Kyma. The tests run daily on Prow to validate if the backup and restore process works for all components.

## Prerequisites

To set up the project, use these tools:

* Version 1.11 or higher of [Go](https://golang.org/dl/)
* Version 0.5 or higher of [Dep](https://github.com/golang/dep)
* The latest version of [Docker](https://www.docker.com/)
* [Velero](../../../resources/backup/README.md#details)

>**NOTE:** Use [these](https://kyma-project.io/docs/master/components/backup) guidelines to configure Velero for a specific storage provider.

## Usage

The backup and restore [continuous integration flow](https://github.com/kyma-project/test-infra/blob/master/prow/scripts/cluster-integration/kyma-gke-backup-test.sh) looks as follows:

1. Create a new Namespace.
2. Create new resources in the Namespace. The resources can be Namespace-scoped or cluster-wide.
3. Verify that the resources work.
4. Back up the Namespace and all the resources in it.
5. Remove the cluster.
6. Re-create the cluster.
7. Restore Kyma from the backup.
8. Verify if the resources contained in the restored cluster work.

### Use environment variables

Use the following environment variables to configure the tests:

| Name | Required | Default | Description |
|-----|:---------:|--------|------------|
| **DOMAIN** | NO | `kyma.local` | The domain where Kyma runs. |
| **USER_EMAIL** | YES | None | The email address for authentication in Dex. |
| **USER_PASSWORD** | YES | None | The password for authentication in Dex. |
| **KUBECONFIG** | NO | None | The path to the `kubeconfig` file needed to run tests outside the cluster. |

### Use flags

Use the following flags to configure the application:

| Name | Required | Description |
|-----|:---------:|------------|
| **action** | YES | Defines what kind of action to execute. The possible values are `testBeforeBackup` and `testAfterRestore`. |

See the example:

```bash
go test restore_cluster_backup_test.go -test.v -action=testBeforeBackup
```

## Development

This section presents how to add and run a new test.

### Add a new test

Add a new test under the `pkg/tests/{domain-name}` directory and implement the following interface:

```go
type BackupTest interface {
    CreateResources(namespace string)
    TestResources(namespace string)
}
```

The functions work as follows:

* The `CreateResources` function installs the required test data before the backup process starts.
* The `TestResources` function validates if the test data works as expected.

After the pipeline executes the backup and restore process, the `TestResources` function validates if the restore worked as expected.

The test creates a new Namespace called `{TestName}-{UUID}`. This Namespace should contain all resources created during the test. If required, the resources can be created in other Namespaces as well.

### Run end-to-end tests locally

> **NOTE:** Before running the test, configure Velero using [these](https://kyma-project.io/docs/master/components/backup) guidelines.

Run tests:

```bash
env KUBECONFIG={KUBECONFIG} go test ./... -count=1 -timeout=0
```

where:

* `{KUBECONFIG}` is the path to the `kubeconfig` file.

### Run tests using the Helm chart

Run the tests using Helm:

1. Prepare the data:

    ```bash
    helm install chart/backup-test/ --name "backup-test" --namespace backup-test --set global.ingress.domainName="$CLUSTER_DOMAIN" --set-file global.adminEmail=<(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.email}" | base64 --decode) --set-file global.adminPassword=<(kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode)
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
env KUBECONFIG={KUBECONFIG} telepresence --run go test ./... -count=1 -timeout=0
```

where:

* `{KUBECONFIG}` is the path to the `kubeconfig` file.

### Verify the code

Before each commit, use the `make verify` command to test your changes

### Project structure

The repository has the following structure:

```text
├── chart                           # The Helm chart for deploying the backup test application
├── pkg                             # The package where backup tests and utility functions are defined. Put your test under /tests folder.
├── backup_test.go                  # The entrypoint for backup test runner
├── Gopkg.toml                      # A dep manifest
└── Gopkg.lock                      # A dep lock which is generated automatically. Do not edit it.
```
