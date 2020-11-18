# Kyma Operator

## Overview

Kyma Operator is a tool for installing all Kyma components. The project is based on the Kubernetes operator pattern. It tracks changes of the Installation custom resource and installs, uninstalls, and updates Kyma accordingly. 

>**NOTE:** The Kyma Operator is designed to work as a [singleton](https://en.wikipedia.org/wiki/Singleton_pattern), meaning that you should deploly only one replica of the controller, and only one Installation CR.

The controller extends the api-server with the following custom resources:
- `installation.installer.kyma-project.io/v1alpha1`

## Prerequisites

- recent version of Go with the support for modules, for example 1.12.6
- make
- kubectl
- access to Kubernetes environment: Minikube or a remote Kubernetes cluster

## Details

### Use commandline flags

| Name | Required | Description | Default value |
|------|:----------:|-------------|-----------------|
| **kubeconfig** | NO | Path to a `kubeconfig` file. | Taken from the **KUBECONFIG** environment parameter.|
| **kymaDir** | NO | Directory holding helm releases within the image or Pod. | `/kyma` |
| **backoffIntervals** | NO | Number of seconds to wait before subsequent retries. | `10,20,40,60,80` |
| **overrideLogFile** | NO | Log file to print installation overrides to. | `/dev/stdout` |
| **overrideLogFormat** | NO | Installation override log format. The accepted values are text or JSON. | `text` |
| **helmMaxHistory**  | NO | Maximum number of releases returned by the Helm release history query. | `10` |
| **helmDriver** | NO | Helm backed storage drivers. Read more about [options and details](https://helm.sh/docs/helm/helm/#helm). | `secrets` |
| **helmDebugMode** | NO | Enables/Disables Helm client output in the Kyma Operator logs. | `false` |

### Supported environment variables

| Name | Default | Description |
| ---- | ------- | ----------- |
| **INST_NAMESPACE** | `default` | Namespace in which the Installation CR is located. |
| **INST_RESOURCE** | `kyma-installation` | Name of the Installation custom resource. |
| **OVERRIDES_NAMESPACE** | `kyma-installer` | Namespace in which the Installer overrides are located. |

## Installer custom resource

The `installations.installer.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format the Kyma Operator Controller listens for. To get the up-to-date CRD and show
the output in the `yaml` format, run this command:

```bash
kubectl get crd installations.installer.kyma-project.io -o yaml
```

### Sample custom resource

The [Installation custom resource file](https://kyma-project.io/docs/root/kyma/#custom-resource-installation) provides the basic information for Kyma installation.

## Additional information

### Upgrade Kyma

The upgrade procedure relies heavily on Helm. When you trigger the upgrade, Helm performs `helm upgrade` on Helm releases that exist in the cluster and are defined in the Kyma version you're upgrading to. If a Helm release is defined in the Kyma version you're upgrading to but is not present in the cluster when you trigger the upgrade, Helm performs `helm install` and installs such a release.

When you trigger the Kyma upgrade, the Kyma Operator lists the Helm releases installed in your current Kyma version. This list is compared against the list of Helm releases of the Kyma version you're upgrading to, to match the releases by their names. The releases that match between versions are upgraded through `helm upgrade`. Releases that don't match are treated as new and installed through `helm install`.

>**NOTE:** If you changed the name of a Helm release for a component, remove it before upgrading Kyma to prevent a situation where two Helm releases of the same component exist in the cluster.

The Operator doesn't migrate custom resources to a new version when the upgrade is triggered. The custom resource backward compatibility, or lack thereof, is determined at the component or Helm release level.
