# Backup and restore E2E tests

## Overview

This project contains end-to-end tests run during Kyma installation on Google Cloud Platform. The tests are written in Go. 

The end-to-end test scenario looks as follows:

- Create a new namespace
- Create a new function in the Namespace.
- Verify that the function works

- Back up the Namespace and all the resources in it.
- Remove the namespace and all its components from the cluster

- Restore the namespace and its content from the backup
- Verify the function contained in the restored namespace works

## Usage

In order to run this end-to-end test just run

```
helm install deploy/chart/backup-test --namespace end-to-end --name backup-test
helm test backup-test
```

The test creates a new namespace called `restore-test-<UUID>`. This namespace contains all resources created during the test.
