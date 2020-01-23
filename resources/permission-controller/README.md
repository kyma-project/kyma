# Permission-controller

## Introduction
This chart bootstraps a permission-controller deployment on a Kubernetes cluster.

## Overview
Permission-controller provides a mechanism for granting admin privileges within custom Namespaces to selected user groups. Under the hood, the component is a Kubernetes controller that watches for instances of `Namespace core/v1` objects and ensures desired RBAC configuration by creating and updating objects of type `rolebindings.rbac.authorization.k8s.io`.

## Installation
Being an integral part of Kyma, permission-controller is available by default in both cluster and local environments. As with the remaining Kyma components, permission-controller is installed using [Helm](https://helm.sh).

## Configuration

See [this](https://kyma-project.io/docs/master/components/security/#configuration-permission-controller-chart) document to learn how to configure the controller.