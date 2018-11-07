# End-to-end backup and restore tests

## Overview

This project contains end-to-end backup and restore tests for resources related to Kyma Environments.

## Usage

See the example usage of running backup phase tests:

```bash
go test -run Backup --kubeconfig $KUBECONFIG --namespace test
```

This test creates a `test` Namespace with several resources, such as Service Instances or bindings, which depends on brokers selected for the test. It also creates two backups: `test-re`, which contains cluster-scoped  RemoteEnvironments, and `test-ns`, which contains Namespace-scoped resources. The backup prefix always contains the name of the Namespace.

 See the example usage of running restore phase tests:

```bash
go test -run Backup --kubeconfig $KUBECONFIG --namespace test
```

This test looks for `test-re` and `test-ns` backups and restores them into a Kubernetes cluster. It also checks Service Instances or bindings for a defined time. If Service Instances or bindings do not reach the `ready` phase, the test fails.
>**NOTE:** The backup name prefix is always the same as the name of the Namespace.
