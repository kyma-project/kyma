---
title: Reinstall Kyma
type: Installation
---

The custom scripts allow you to remove Kyma from a Minikube cluster and reinstall Kyma without removing the cluster.

> **NOTE:** These scripts do not delete the cluster from your Minikube. This allows you to quickly reinstall Kyma.

1. Use the `Kyma CLI` uninstall Kyma from the cluster. Run:
  ```bash
  kyma uninstall
  ```

2. Run this command to reinstall Kyma on an existing cluster:
  ```bash
  kyma install
  ```
