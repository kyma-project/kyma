---
title: Reinstall Kyma
type: Installation
---

The Kyma CLI allow you to remove Kyma from a Minikube cluster and reinstall Kyma without removing the cluster.

> **NOTE:** The Kyma CLI can uninstall Kyma without deleting the cluster from your Minikube. This allows you to quickly reinstall Kyma.

1. Use Kyma CLI to uninstall Kyma from the cluster. Run:
  ```bash
  kyma uninstall
  ```

2. Run this command to reinstall Kyma on an existing cluster:
  ```bash
  kyma install
  ```
