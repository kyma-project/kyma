# Kyma Operator

## Overview

Kyma Operator is a tool for installing all Kyma components. The project is based on the Kubernetes operator pattern. It tracks changes of the Installation custom resource and installs, uninstalls, and updates Kyma accordingly.

## Prerequisites

- Minikube 0.26.1
- kubectl 1.9.0
- Docker
- jq

## Development

Before each commit, use the [`Makefile`](./Makefile) script to test your changes:
  ```bash
  make verify
  ```

## Build a Docker image

Run the [`build.sh`](./scripts/build.sh) script to build a Docker image of the Kyma Operator:
  ```
  ./scripts/build.sh
  ```

## Run on Minikube using local sources

Export the path to Kyma sources as an environment variable. This environment variable is used to trigger shell scripts located in Kyma.
  ```
  export KYMA_PATH={PATH_TO_KYMA_SOURCES}
  ```

Use this script to run Minikube, set up the Kyma Operator, and install Kyma from local sources. If Minikube was started before, it will be restarted.
  ```
  ./scripts/run.sh --local --cr {PATH_TO_CR}
  ```

See the [Custom resource file](#custom-resource-file) section in this document to learn how to generate a custom resource file.

To track progress of the installation, run:
  ```
  ../../installation/scripts/is-installed.sh
  ```

## Rerun Kyma without restarting Minikube

This scenario is useful when you want to reuse cached Docker images.

Run this script to clear running Minikube:
  ```
  ../../installation/scripts/clean-up.sh
  ```

Execute the `run.sh` script with the `--skip-minikube-start` flag to rerun Kyma installation without stopping your Minikube:
  ```
  ./scripts/run.sh --skip-minikube-start
  ```

## Upgrade Kyma

The upgrade procedure relies heavily on Helm. When you trigger the upgrade, Helm performs `helm upgrade` on Helm releases that exist in the cluster and are defined in the Kyma version you're upgrading to. If a Helm release is defined in the Kyma version you're upgrading to but is not present in the cluster when you trigger the upgrade, Helm performs `helm install` and installs such a release.

When you trigger Kyma upgrade, the Kyma Operator lists the Helm releases installed in your current Kyma version. This list is compared against the list of Helm releases of the Kyma version you're upgrading to. The releases are matched by their names. The releases that match between versions are upgraded through `helm upgrade`. Releases that don't match are treated as new and installed through `helm install`.

>**NOTE:** If you changed the name of a Helm release for a component, remove it before upgrading Kyma to prevent a situation where two Helm releases of the same component exist in the cluster.

The Operator doesn't support rollbacks.

The Operator doesn't migrate Custom Resources to a new version when update is triggered. Custom Resource backward compatibility, or lack thereof, is determined at the component or Helm release level.

## Update the Kyma cluster

Connect to the cluster that hosts your Kyma installation. Prepare the URL to the updated Kyma `tar.gz` package. Run the following command to edit the installation CR:
  ```
  kubectl edit installation/{CR_NAME}
  ```

Change the `url` property in **spec** to `{URL_TO KYMA_TAR_GZ}`. Trigger the update by overriding the **action** label in CR:
  ```
  kubectl -n default label installation/{CR_NAME} action=install --overwrite
  ```

## Uninstall Kyma

Run this command to uninstall Kyma:
  ```
  kubectl -n default label installation/kyma-installation action=uninstall
  ```

>**NOTE:** Uninstallation is an experimental feature that invokes the `helm delete --purge` command for each release specified in the CR. It is Helm's policy to prevent some types of resources, such as CRDs or Namespaces, from being deleted. Delete them manually before you attempt to install Kyma again on a cluster from which it was uninstalled.  

## Custom resource file

The [Installation custom resource file](https://kyma-project.io/docs/root/kyma/#custom-resource-installation) provides the basic information for Kyma installation.

The required properties include:

- `url` which is the URL address to a Kyma charts package. Only `tar.gz` is supported and it is required for non-local installations only.
- `version` which is the version of Kyma.
- `components` which is the list of Kyma components.

## Generate the custom resource file

Generate a custom resource file using the [`create-cr.sh`](../../installation/scripts/create-cr.sh) script. It accepts the following arguments:

- `--output` is a mandatory parameter which indicates the location of the custom resource output file.
- `--url` is the URL to the Kyma package.
- `--version` is the version of Kyma.

For example:
  ```
  ../../installation/scripts/create-cr.sh --output installer-cr.yaml --url {URL_TO_KYMA_TAR_GZ} --version 0.8.0
  ```

## Static overrides for cluster installations

You can define cluster-specific overrides for each root chart. In the cluster deployment scenario, the Kyma Operator reads the `cluster.yaml` file in each root chart and appends its content to the overrides applied during the Helm installation.
