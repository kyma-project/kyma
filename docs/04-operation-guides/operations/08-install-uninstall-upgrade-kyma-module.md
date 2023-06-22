---
title: Install, uninstall and upgrade a Kyma module
---

The Kyma project is currently in the transition phase from classic to modular Kyma. Learn how to install, uninstall and upgrade a Kyma module. To see the list of all Kyma modules, go to [Overview](/docs/01-overview/README.md).

## Install a Kyma module

To install a module, deploy its module manager and apply the module configuration. The table lists the available Kyma modules and provides quick installation steps. For more details, see the module documentation in GitHub.

<table>
<tr>
<td> <b>Module</b> </td> <td> <b>Installation steps</b> </td> <td> <b>Documentation</b> </td>
</tr>
<tr>
<td> Keda </td>
<td>

```bash
kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-manager.yaml
kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-default-cr.yaml
```

</td>
<td> <a href="https://github.com/kyma-project/keda-manager">Keda Manager</a></td>
</tr>
<tr>
<td> BTP Operator </td>
<td>

```bash
kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-operator.yaml
kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btpoperator-default-cr.yaml
```

<td> <a href="https://github.com/kyma-project/btp-manager">BTP Manager</a></td>
</td>
</tr>
</table>

## Uninstall a Kyma module

To uninstall a Kyma module, use the `kubectl delete` command. First, delete the module configuration, and then the module manager. Use the paths from the table in the [Install a Kyma module](#install-a-kyma-module) section. Run:

```bash
kubectl delete {PATH_TO_THE_MODULE_CUSTOM_RESOURCE}
kubectl delete {PATH_TO_THE_MODULE_MANAGER_YAML_FILE}
```

> **TIP:** Before you delete the module manager, wait for the module custom resource deletion to be complete to avoid leaving some resources behind.

## Upgrade a Kyma module

To upgrade a Kyma module to the latest version, run the same `kubectl` commands used for its installation.
