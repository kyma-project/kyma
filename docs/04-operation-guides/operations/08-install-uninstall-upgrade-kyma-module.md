---
title: Install, uninstall and upgrade a Kyma module
---

The Kyma project is currently in the transition phase from classic to modular Kyma. Learn how to install, uninstall and upgrade a Kyma module. To see the list of all Kyma modules, go to [Overview](/docs/01-overview/README.md).

## Install a Kyma module

To install a module, deploy its module manager and apply the module configuration. The table lists the available Kyma modules and provides quick installation steps. For more details, see the module documentation in GitHub.

| Module       | Installation steps               | Documentation     |
|--------      |----------------------------------|-------------------|
| Keda         | <pre lang="bash">kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-manager.yaml<br>kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-default-cr.yaml</pre>| [Keda Manager](https://github.com/kyma-project/keda-manager) |
| BTP Operator | <pre lang="bash">kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-operator.yaml<br>kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btpoperator-default-cr.yaml</pre> | [BTP Manager](https://github.com/kyma-project/btp-manager) |

## Uninstall a Kyma module

To uninstall a Kyma module, use the `kubectl delete` command. Delete the module configuration and the module manager. Use paths from the table in the [Install a Kyma module section](#install-a-kyma-module). Run:

```bash
kubectl delete {PATH_TO_THE_MODULE_CUSTOM_RESOURCE}
kubectl delete {PATH_TO_THE_MODULE_MANAGER_YAML_FILE}
```

> **TIP:** Before you delete the module manager, wait for the module custom resource deletion to be complete to avoid leaving some resources behind.

## Upgrade a Kyma module

To upgrade a Kyma module to the latest version, run the same `kubectl` commands used for its installation.
