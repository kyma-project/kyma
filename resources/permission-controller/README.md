# Permission-controller

## Introduction
This chart bootstraps a permission-controller deployment on a Kubernetes cluster using [Helm](https://helm.sh/).

## Overview
Permission-controller provides a mechanism for granting admin privileges within custom Namespaces to selected user groups. Under the hood, the component is a Kubernetes controller that watches for instances of `Namespace core/v1` objects and ensures desired RBAC configuration by creating and updating objects of type `rolebindings.rbac.authorization.k8s.io`.

## Installation
Being an integral part of Kyma, permission-controller is available by default in both cluster and local environments.

## Configuration

The following table lists the configurable parameters of the permission-controller chart and their default values. As with the remaining Kyma components, permission-controller is installed using [Helm](https://helm.sh) .

| Parameter | Type | Description | Default value |
| --------- | ---- | ----------- | ------------- |
| `config.subjectGroups` | []string | List of user groups whose members are to be granted admin privileges in all Namespaces, except those specified in `config.namespaceBlacklist`. | ["namespace-admins"] |
| `config.namespaceBlacklist` | []string | List of Namespaces that will not be accessible by members of the user groups specified in `config.subjectGroups`. | ["kyma-system, istio-system, default, knative-eventing, knative-serving, kube-node-lease, kube-public, kube-system, kyma-installer, kyma-integration, natss"] |
| `config.enableStaticUser`| boolean | Determine whether to grant admin privileges to the `namespace.admin@kyma.cx` static user. | true |

Configure the controller by modifying the [`values.yaml`](./values.yaml) file or by creating a [Configmap with Helm overrides](https://kyma-project.io/docs/#configuration-helm-overrides-for-kyma-installation).