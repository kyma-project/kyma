# Permission-controller

## Introduction
This chart bootstraps a permission-controller deployment on a Kubernetes cluster.

## Overview
The Permission Controller listens for new Namespaces and creates a RoleBinding for the users of specified groups to the **kyma-admin** role within these Namespaces.  The Controller uses a blacklist mechanism, which defines the Namespaces in which the users of the defined groups are not assigned the **kyma-admin** role. When the Controller is deployed in a cluster, it checks all existing Namespaces and assigns the roles accordingly.

## Installation
Being an integral part of Kyma, permission-controller is available by default in both cluster and local environments. As with the remaining Kyma components, permission-controller is installed using [Helm](https://helm.sh).

## Configuration

See [this](https://kyma-project.io/docs/master/components/security) document to learn how to configure the controller.
