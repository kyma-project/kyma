# Installer

## Overview

Kyma Installer is a tool for installing all Kyma components. The project is based on the Kubernetes operator pattern. It tracks changes of the Installation custom resource and installs, uninstalls, and updates Kyma accordingly.

## Prerequisites

- Minikube 0.26.1
- kubectl 1.9.0
- Docker
- jq

## Development

Before each commit, use the [`before-commit.sh`](./before-commit.sh) script to test your changes:
  ```
  ./before-commit.sh
  ```

## Build a Docker image

Run the [`build.sh`](./scripts/build.sh) script to build a Docker image of the Installer:
  ```
  ./scripts/build.sh
  ```

## Run on Minikube using local sources

Export the path to Kyma sources as an environment variable. This environment variable is used to trigger shell scripts located in Kyma.
  ```
  export KYMA_PATH={PATH_TO_KYMA_SOURCES}
  ```

Use this script to run Minikube, set up the Installer, and install Kyma from local sources. If Minikube was started before, it will be restarted.
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

## Update the Kyma cluster

Connect to the cluster that hosts your Kyma installation. Prepare the URL to the updated Kyma `tar.gz` package. Run the following command to edit the installation CR:
  ```
  kubectl edit installation/{CR_NAME}
  ```

Change the `url` property in **spec** to `{URL_TO KYMA_TAR_GZ}`. Trigger the update by overriding the **action** label in CR:
  ```
  kubectl label installation/{CR_NAME} action=install --overwrite
  ```

## Uninstall Kyma

Run this command to uninstall Kyma:
  ```
  kubectl label installation/kyma-installation action=uninstall
  ```

>**NOTE:** Uninstallation is an experimental feature that invokes the `helm delete --purge` command for each release specified in the CR. It is Helm's policy to prevent some types of resources (e.g. CRDs or namespaces) from being deleted. Make sure to remove them manually before you attempt to install Kyma again on a cluster from which it was uninstalled.  

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

You can define cluster-specific overrides for each root chart. In the cluster deployment scenario, the Installer reads the `cluster.yaml` file in each root chart and appends its content to the overrides applied during the Helm installation.
