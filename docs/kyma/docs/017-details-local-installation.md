---
title: Local installation
type: Details
---

For better understanding of complex installation process, this guide describes it step by step.

## Start local installation

To fire local installation run following command:
```
./installation/cmd/run.sh
```

This script sets up default parameters, starts Minikube, builds Kyma-Installer, generates local configuration, creates `installer` Custom Resource and sets up Installer. Subsequent sections provide a detailed description of each step.

### Installation parameters

You can execute the `installation/cmd/run.sh` script with the following parameters:

- `--skip-minikube-start` - it skips execution of the `installation/scripts/minikube.sh` script. See the "Start Minikube" section for more details.
- `--vm-driver` -  either `virtualbox` or `hiperkit` depending on your operating system

## Start Minikube

> **NOTE:** To work with Kyma, use only the provided scripts and commands. Kyma does not work on a basic Minikube cluster that you can start using the `minikube start` command or stop with the `minikube stop` command. If you don't need Kyma on Minikube anymore, remove the cluster with the `minikube delete` command.

Purpose of `installation/scripts/minikube.sh` is to configure and start Minikube. The script also checks if your development environment is cofigured to handle Kyma installation. This includes checking Minikube and Kubectl versions. In case Minikube is already initialized you will be prompted to agree to remove previous Minikube cluster. Script exits if you don't want to restart your cluster.

>**NOTE:** Minikube is configured to disable default nginx ingress controller.

>**NOTE:** For the complete list of parameters passed to `minikube start` command, refer to the `installation/scripts/minikube.sh` script.

Once Minikube is up and running, the script adds local installation entries to /etc/hosts.

## Build Kyma-Installer

Installer is application based on a [Kubernetes operator](https://coreos.com/operators/). Its purpose is to install Helm charts defined in the Installer Custom Resource. Kyma-Installer is a Docker image that bundles Installer binary with Kyma charts. 

`installation/scripts/build-kyma-installer.sh` script extracts Kyma-Installer image name from the `installer.yaml` deployment file and uses it to build a Docker image inside Minikube. This image will contain local Kyma sources from `resources` folder. 

>**NOTE:** For Kyma-Installer Docker image details refer to the `kyma-installer/kyma.Dockerfile` file.

## Generate local configuration for Azure-Broker

The Azure-Broker sub-component is a part of the `core` deployment that provisions managed services in the Microsoft Azure cloud. To enable Azure-Broker, export the following environment variables:
 - AZURE_BROKER_SUBSCRIPTION_ID
 - AZURE_BROKER_TENANT_ID
 - AZURE_BROKER_CLIENT_ID
 - AZURE_BROKER_CLIENT_SECRET

>**NOTE:** As the Azure credentials are converted to a Kubernetes Secret, make sure the exported values are base64-encoded.

## Create installer custom resource

`installation/scripts/create-cr.sh` script prepares Installer Custom Resource from `installation/resources/installer-cr.yaml.tpl` template. In local installation scenario default Installer Custom Resource will be used. `kyma-installer` already contains local Kyma resources bundled, thus `url` is ignored by Installer component. 

>**NOTE:** For Installer Custom Resource details refer to the `docs/kyma/docs/040-cr-installation.md` file.

## Kyma-Installer deployment

The `installation/scripts/installer.sh` script creates the default RBAC Role, installs [Tiller] (https://docs.helm.sh/), and deploys the Kyma-Installer component.

>**NOTE:** For Kyma-Installer deployment details refer to the `installation/resources/installer.yaml` file.

The script applies the Installer Custom Resource and marks it with the `action=install` Label, which triggers the Kyma installation.

>**NOTE:** Kyma installation runs in background. Execute `./installation/scripts/is-installed.sh` to follow the installation process.
