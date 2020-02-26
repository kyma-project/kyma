# Kyma Operator

## Overview

Kyma Operator is a tool for installing all Kyma components. The project is based on the Kubernetes operator pattern. It tracks changes of the Installation custom resource and installs, uninstalls, and updates Kyma accordingly.

## Prerequisites

- Docker

## Development

Before each commit, use the [`Makefile`](./Makefile) script to test your changes:
  ```bash
  make verify
  ```

## Rebuild Kubernetes client libraries

After introducing changes to the `apis/installer/v1alpha1` package, you have to regenerate client libraries by running:
  ```
  sh ./hack/update-codegen.sh
  ```

## Build a Docker image

Use the [`Makefile`](./Makefile) script to build a Docker image of the Kyma Operator:
  ```
  export DOCKER_PUSH_REPOSITORY={DOCKER_REPOSITORY}
  make build-image
  ```

## Details

### Upgrade Kyma

The upgrade procedure relies heavily on Helm. When you trigger the upgrade, Helm performs `helm upgrade` on Helm releases that exist in the cluster and are defined in the Kyma version you're upgrading to. If a Helm release is defined in the Kyma version you're upgrading to but is not present in the cluster when you trigger the upgrade, Helm performs `helm install` and installs such a release.

When you trigger Kyma upgrade, the Kyma Operator lists the Helm releases installed in your current Kyma version. This list is compared against the list of Helm releases of the Kyma version you're upgrading to. The releases are matched by their names. The releases that match between versions are upgraded through `helm upgrade`. Releases that don't match are treated as new and installed through `helm install`.

>**NOTE:** If you changed the name of a Helm release for a component, remove it before upgrading Kyma to prevent a situation where two Helm releases of the same component exist in the cluster.

The Operator doesn't support rollbacks.

The Operator doesn't migrate Custom Resources to a new version when update is triggered. Custom Resource backward compatibility, or lack thereof, is determined at the component or Helm release level.

### Custom resource file

The [Installation custom resource file](https://kyma-project.io/docs/root/kyma/#custom-resource-installation) provides the basic information for Kyma installation.
