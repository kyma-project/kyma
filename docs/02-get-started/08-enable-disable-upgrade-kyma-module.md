---
title: Enable, disable and upgrade a Kyma module
---

Learn how to enable, disable and upgrade a Kyma module. To see the list of all available and planned Kyma modules, go to [Overview](../../01-overview/README.md).

## Enable a Kyma module

To enable a module, deploy its module manager and apply the module configuration. See the already available Kyma modules with their quick installation steps and links to their GitHub repositories:

### Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- Kubernetes cluster, or [k3d](https://k3d.io) for local installation

### Steps

#### Keda

```bash
kubectl create ns kyma-system
kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-manager.yaml
kubectl apply -f https://github.com/kyma-project/keda-manager/releases/latest/download/keda-default-cr.yaml -n kyma-system
```

For more details see, the [`keda-manager`](https://github.com/kyma-project/keda-manager) repository in GitHub.

#### BTP Operator

```bash
kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-manager.yaml
kubectl apply -f https://github.com/kyma-project/btp-manager/releases/latest/download/btp-operator-default-cr.yaml
```

> **CAUTION:** The CR is in the `Warning` state and the message is `Secret resource not found reason: MissingSecret`. To create a Secret, follow the instructions in the [`btp-manager`](https://github.com/kyma-project/btp-manager/blob/main/docs/user/02-10-usage.md#create-and-install-secret) repository.

#### Serverless

```bash
kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-manager.yaml
kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-default-cr.yaml
```

#### Telemetry

```bash
kubectl create ns kyma-system
kubectl apply -f https://github.com/kyma-project/telemetry-manager/releases/latest/download/rendered.yaml
kubectl apply -f https://github.com/kyma-project/telemetry-manager/releases/latest/download/telemetry-default-cr.yaml -n kyma-system
```

For more installation options, see the [installation instruction](https://github.com/kyma-project/telemetry-manager/blob/main/docs/contributor/installation.md) in the module repository.

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
