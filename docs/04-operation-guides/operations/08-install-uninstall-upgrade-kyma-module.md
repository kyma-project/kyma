---
title: Install, uninstall and upgrade a Kyma module
---

The Kyma project is currently in the transition phase from classic to modular Kyma. Learn how to install, uninstall and upgrade a Kyma module. To see the list of all Kyma modules, go to [Overview](/docs/01-overview/README.md).

## Install a Kyma module

To install a module, deploy its module manager and apply the module configuration. See the available Kyma modules with quick installation steps and links to their GitHub repositories.

### Keda

```bash
kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-manager.yaml
kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-default-cr.yaml
```

For more details see, the [`keda-manager`](https://github.com/kyma-project/keda-manager) repository in GitHub.

### BTP Operator

```bash
kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/rendered.yaml
kubectl apply -f https://raw.githubusercontent.com/kyma-project/btp-manager/main/config/samples/operator_v1alpha1_btpoperator.yaml
```

For more details see, the [`btp-manager`](https://github.com/kyma-project/btp-manager) repository in GitHub.

## Uninstall a Kyma module

To uninstall a Kyma module, use the `kubectl delete` command. First, delete the module configuration, and then the module manager. Use the paths from the table in the [Install a Kyma module](#install-a-kyma-module) section. Run:

```bash
kubectl delete {PATH_TO_THE_MODULE_CUSTOM_RESOURCE}
kubectl delete {PATH_TO_THE_MODULE_MANAGER_YAML_FILE}
```

> **TIP:** Before you delete the module manager, wait for the module custom resource deletion to be complete to avoid leaving some resources behind.

## Upgrade a Kyma module

To upgrade a Kyma module to the latest version, run the same `kubectl` commands used for its installation.
