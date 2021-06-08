---
title: Installation troubleshooting
type: Troubleshooting
---

## Kyma Installer doesn't respond as expected

If the Installer does not respond as expected, check the installation status using the `is-installed.sh` script with the `--verbose` flag added. Run:

```bash
scripts/is-installed.sh --verbose
```

## Installation successful, component not working

If the installation is successful but a component does not behave in the expected way, inspect Helm releases for more details on all of the installed components.

Run this command to list all of the available Helm releases:

```bash
helm list --all-namespaces --all
```

Run this command to get more detailed information about a given release:

```bash
helm status {RELEASE_NAME} -n {RELEASE_NAMESPACE}
```

>**NOTE:** Names of Helm releases correspond to names of Kyma components.


Additionally, see if all deployed Pods are running. Run this command:

```bash
kubectl get pods --all-namespaces
```

The command retrieves all Pods from all Namespaces, the status of the Pods, and their instance numbers. Check if the status is `Running` for all Pods. If any of the Pods that you require do not start successfully, install Kyma again.