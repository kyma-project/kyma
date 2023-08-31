---
title: Install, uninstall and upgrade Kyma with a module
---

This guide shows how to quickly install, uninstall or upgrade Kyma with specific modules. To see the list of all available and planned Kyma modules, go to [Kyma modules](../README.md#kyma-modules).

> **NOTE:** This guide describes installation of standalone Kyma with specific modules, and not how to enable modules in SAP BTP, Kyma runtime (SKR).

## Install Kyma with a module

To install a module, deploy its module manager and apply the module configuration. The operation installs a Kyma module of your choice on a Kubernetes cluster. See the already available Kyma modules with their quick installation steps and links to their GitHub repositories:

### Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- Kubernetes cluster, or [k3d](https://k3d.io) (v5.x or higher) for local installation
- `kyma-system` Namespace created

### Steps

#### [Keda](https://github.com/kyma-project/keda-manager)

```bash
kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-manager.yaml
kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-default-cr.yaml -n kyma-system
```

#### [BTP Operator](https://github.com/kyma-project/btp-manager)

```bash
kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-manager.yaml
kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-operator-default-cr.yaml -n kyma-system
```

> **CAUTION:** The CR is in the `Warning` state and the message is `Secret resource not found reason: MissingSecret`. To create a Secret, follow the instructions in the [`btp-manager`](https://github.com/kyma-project/btp-manager/blob/main/docs/user/02-10-usage.md#create-and-install-secret) repository.

#### [Serverless](https://github.com/kyma-project/serverless-manager)

```bash
kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default_serverless_cr.yaml  -n kyma-system
```

#### [Telemetry](https://github.com/kyma-project/telemetry-manager)

```bash
kubectl apply -f https://github.com/kyma-project/telemetry-manager/releases/latest/download/rendered.yaml
kubectl apply -f https://github.com/kyma-project/telemetry-manager/releases/latest/download/telemetry-default-cr.yaml -n kyma-system
```

## Uninstall Kyma with a module

You uninstall Kyma with a module with the `kubectl delete` command.

1. Find out the paths for the module you want to disable; for example, from the [Install Kyma with a module](#install-kyma-with-a-module) section.

2. Delete the module configuration:

   ```bash
   kubectl delete {PATH_TO_THE_MODULE_CUSTOM_RESOURCE}
   ```

3. To avoid leaving some resources behind, wait for the module custom resource deletion to be complete.

4. Delete the module manager:

   ```bash
   kubectl delete {PATH_TO_THE_MODULE_MANAGER_YAML_FILE}
   ```

## Upgrade a Kyma module

To upgrade a Kyma module to the latest version, run the same `kubectl` commands used for its installation.
