# End-to-end Backp and Restore Tests

## Overview


This project contains end-to-end upgrade tests for the Kyma installation to validate if the restore process works for all components. The tests run daily on Prow.
 

## Prerequisites

Install [Ark](../../../resources/ark/README.md#details) and [configure](../../../docs/backup/docs/03-01-backup-configuration.md) it for a specific storage provider.


## Usage

The backup and restore [continuous integration flow](https://github.com/kyma-project/test-infra/blob/master/prow/scripts/cluster-integration/kyma-gke-end-to-end-test.sh) looks as follows:

1. Create a new Namespace.
2. Create new resources in the Namespace. The resources can be Namespace-wide or cluster-wide.
3. Verify that the resources work.
4. Back up the Namespace and all the resources in it.
5. Remove the Namespace and all its resources.
6. Restore the Namespace and its resources from the backup.
7. Verify if the resources contained in the restored Namespace work.


## Development

This section presents how to add and run a new test. 

### Add a new test

Add a new test under the `backup-restore/backupe2e/{domain-name}` directory and implement the following interface:

```go
type BackupTest interface {
    CreateResources(namespace string)
    TestResources(namespace string)
    DeleteResources()
}
```
The functions work as follows:

- The `TestResources` function validates if the test data works as expected. 
- The `CreateResources` function installs the required test data before the backup process starts.
- The `DeleteResources` function deletes the resources that are a part of the cluster before executing the test. The resources need to be deleted to test the restore process.

After the pipeline executes the backup and restore processes on the cluster, the `TestResources` function validates if the restore worked as expected.

### Run end-to-end tests locally

> **NOTE:** Before running the test, configure Ark using [these](../../../docs/backup/docs/03-01-backup-configuration.md) guidelines.

Run tests:
```bash
env KUBECONFIG=/Users/$User/.kube/config go run restore_cluster_backup_test.go --action executeTests`
```

### Run tests using a Helm chart

Run the application using Helm:

Prepare the data:

```bash
helm install deploy/chart/backup-test --namespace end-to-end --name backup-test
```
Run tests:

```bash
helm test backup-test
```
The test creates a new Namespace called `restore-test-<UUID>`. This Namespace contains all resources created during the test.

### Run tests using Telepresence
[Telepresence](https://www.telepresence.io/) allows you to run tests locally while connecting a service to a remote Kubernetes cluster. It is helpful when the test needs access to other services in a cluster.

1. [Install Telepresence](https://www.telepresence.io/reference/install).
2. Run tests:
```bash
env KUBECONFIG=/Users/$User/.kube/config  telepresence --run go run restore_cluster_backup_test.go  --action execute
```