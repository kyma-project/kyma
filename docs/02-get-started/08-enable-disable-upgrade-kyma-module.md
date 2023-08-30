---
title: Enable, disable and upgrade a Kyma module
---

This guide shows how to quickly install, uninstall or remove Kyma with specific modules. To see the list of all available and planned Kyma modules, go to the [Kyma modules](../README.md#kyma-modules).

> **NOTE:** This guide describes installation of a standalone Kyma with specific modules, and not how to enable modules in SAP BTP, Kyma runtime (SKR).

## Enable a Kyma module

To enable a module, deploy its module manager and apply the module configuration. See the already available Kyma modules with their quick installation steps and links to their GitHub repositories:

### Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- Kubernetes cluster, or [k3d](https://k3d.io) for local installation
- `kyma-system` Namespace created

### Steps

#### Keda

```bash
kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-manager.yaml
kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-default-cr.yaml -n kyma-system
```

#### BTP Operator

```bash
kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-manager.yaml
kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-operator-default-cr.yaml -n kyma-system
```

> **CAUTION:** The CR is in the `Warning` state and the message is `Secret resource not found reason: MissingSecret`. To create a Secret, follow the instructions in the [`btp-manager`](https://github.com/kyma-project/btp-manager/blob/main/docs/user/02-10-usage.md#create-and-install-secret) repository.

#### Serverless

```bash
kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default_serverless_cr.yaml  -n kyma-system
```

#### Telemetry

```bash
kubectl apply -f https://github.com/kyma-project/telemetry-manager/releases/latest/download/rendered.yaml
kubectl apply -f https://github.com/kyma-project/telemetry-manager/releases/latest/download/telemetry-default-cr.yaml -n kyma-system
```

## Disable a Kyma module

You disable a Kyma module with the `kubectl delete` command.

1. Find out the paths for the module you want to disable; for example, from the [Enable a Kyma module](#enable-a-kyma-module) section.

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
