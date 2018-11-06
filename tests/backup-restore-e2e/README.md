# Backup/Restore e2e tests

## Overview

This project contains end-to-end tests for backup & restore of resources related to kyma's environments (namespaces).

## Usage

Example usage to run backup phase tests:

```bash
go test -run Backup --kubeconfig $KUBECONFIG --namespace test
```

It will create namespace `test` with several resources (service instances/bindings depends on selected brokers for test) and will create 2 backups: `test-re` (containing cluster scoped resources: RemoteEnvironments) and `test-ns` (containing namespace scoped resources). Backup prefix will always contain namespace name.

Example usage to run restore phase tests:

```bash
go test -run Backup --kubeconfig $KUBECONFIG --namespace test
```

It will look for 2 backups `test-re` and `test-ns` (backup name prefix is always name of namespace) and restore them into k8s cluster and check service instantes/bindings for defined time. If service instances/bindings won't get ready phase the test will fail.