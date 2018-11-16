# Installer

## Overview

Installer is a tool for installing all Kyma components.
The project is based on the Kubernetes operator pattern. It tracks changes of the `Installation` type instance of the custom resource. It also installs, uninstalls, and updates Kyma accordingly.

## Prerequisites

- Minikube 0.26.1
- kubectl 1.9.0
- Docker
- jq

## Development

Before each commit, use the [before-commit.sh](./before-commit.sh) script to test your changes:
```
./before-commit.sh
```

### Build a Docker image

Run the [build.sh](./scripts/build.sh) script to build a Docker image:

```
./scripts/build.sh
```

### Run on Minikube

#### Run with local Kyma resources
```
export KYMA_PATH={PATH_TO_KYMA_SOURCES}
```
This environment variable is used to trigger shell scripts located in Kyma.
```
./scripts/run.sh --local --cr {PATH_TO_CR}
```

It will run Minikube, set up the installer, and install Kyma from local sources. If Minikube was started before, it will be restarted.

See the [Custom resource file](#custom-resource-file) section in this document to learn how to generate a custom resource file.

To track progress of the installation, run:

```
../../installation/scripts/is-installed.sh
```

#### Rerun without restarting Minikube

This scenario is useful when you want to reuse cached Docker images.

Run the following script to clear running Minikube:
```
../../installation/scripts/clean-up.sh
```

Execute the `run.sh` script with the `--skip-minikube-start` flag to rerun Kyma installation without stopping your Minikube:
```
./scripts/run.sh --skip-minikube-start
```

### Update the Kyma cluster

Connect to the cluster that hosts your Kyma installation. Prepare the URL to the updated Kyma `tar.gz` package. Run the following command to edit the installation CR:
```
kubectl edit installation/{CR_NAME}
```
Change the `url` property in **spec** to `{URL TO KYMA TAR GZ}`. Trigger the update by overriding the **action** label in CR:
```
kubectl label installation/{CR_NAME} action=install --overwrite
```

### Update the local Kyma installation

Prepare local changes in Kyma sources. Run the following command to copy the updated sources to the Installer Pod and trigger the update action:
```
../../installation/scripts/update.sh --local --cr-name {CR_NAME}
```

> **NOTE:** You do not have to restart Minikube.

### Uninstall Kyma

Run the following command to completely uninstall Kyma:
```
kubectl label installation/kyma-installation action=uninstall
```

### Custom resource file

The custom resource file for installer provides the basic information for Kyma installation.

The required properties iclude:

- `url` which is the URL address to a Kyma charts package. Only `tar.gz` is supported and it is required for non-local installations only.
- `version` which is the version of Kyma.
- `components` which is the list of Kyma components.


### Generate the custom resource file for installer

Generate a custom resource file using the [create-cr.sh](../../installation/scripts/create-cr.sh) script. It accepts the following arguments:

- `--output` is a mandatory parameter which indicates the location of the custom resource output file.
- `--url` is the URL to the Kyma package.
- `--version` is the version of Kyma.

Example:
```
../../installation/scripts/create-cr.sh --output installer-cr.yaml --url {URL TO KYMA TAR GZ} --version 0.0.1
```

### Enable the Azure Broker

To run the Kyma with the Azure Broker enabled, mark the `azure-broker` subcomponent as enabled either using the `manage-component.sh` script or manually. Specify the Azure credentials as the environment variables providing the following variables along with their values encoded as base64 strings: `AZURE_BROKER_SUBSCRIPTION_ID`, `AZURE_BROKER_TENANT_ID`, `AZURE_BROKER_CLIENT_ID`, and `AZURE_BROKER_CLIENT_SECRET`.

Example:
```
$ export AZURE_BROKER_SUBSCRIPTION_ID="..."
$ export AZURE_BROKER_TENANT_ID="..."
$ export AZURE_BROKER_CLIENT_ID="..."
$ export AZURE_BROKER_CLIENT_SECRET="..."

$ ../../installation/scripts/manage-component.sh azure-broker true
```

## Static overrides for cluster installations

You can define cluster-specific overrides for each root chart. In the cluster deployment scenario, the installer reads the `cluster.yaml` file in each root chart and appends its content to the overrides applied during the
Helm installation.
