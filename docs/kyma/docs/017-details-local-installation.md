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

This script sets up default parameters, starts minikube, builds kyma-installer, generates local configuration, creates `installer` Custom Resource and sets up installer. Subsequent sections provide a detailed description of each step.

### Installation parameters

You can execute the `installation/cmd/run.sh` script with the following parameters:

- `--skip-minikube-start` - it skips execution of the `installation/scripts/minikube.sh` script. See the "Start Minikube" section for more details.
- `--vm-driver` -  either `virtualbox` or `hiperkit` depending on your operating system

## Start minikube

> **NOTE:** To work with Kyma, use only the provided scripts and commands. Kyma does not work on a basic Minikube cluster that you can start using the `minikube start` command or stop with the `minikube stop` command. If you don't need Kyma on Minikube anymore, remove the cluster with the `minikube delete` command.

Purpose of `installation/scripts/minikube.sh` is to configure and start minikube. The script also checks if your development environment is cofigured to handle kyma installation. This includes checking minikube and kubectl versions. In case minikube is already initialized you will be prompted to agree to remove previous minikube cluster. Script exits if you don't want to restart your cluster.

>**NOTE:** Minikube is configured to disable default nginx ingress controller.

>**NOTE:** For the complete list of parameters passed to `minikube start` command, refer to the `installation/scripts/minikube.sh` script.

Once minikube is up and running, the script adds develop domains to /ets/hosts.

## Build kyma-installer

Kyma installs locally using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/).

`installer` is application based on a [Kubernetes operator](https://coreos.com/operators/). Its purpose is to install helm charts defined in the `installer` Custom Resource. `kyma-installer` is `installer` with bundled kyma charts. 

`installation/scripts/build-kyma-installer.sh` script extracts `kyma-installer` image name from the `installer.yaml` deployment file and uses it to build a Docker image inside minikube. This image will contain local kyma sources from `resources` folder. 

>**NOTE:** For `kyma-installer` Docker image details refer to the `kyma-installer/kyma.Dockerfile` file.

## Generate local configuration for Azure-Broker

The `Azure-Broker` sub-component is a part of the `core` deployment that provisions managed services in the Microsoft Azure cloud. To enable `Azure-Broker`, export the following environment variables:
 - AZURE_BROKER_SUBSCRIPTION_ID
 - AZURE_BROKER_TENANT_ID
 - AZURE_BROKER_CLIENT_ID
 - AZURE_BROKER_CLIENT_SECRET

>**NOTE:** As the Azure credentials are converted to a Kubernetes Secret, make sure the exported values are base64-encoded.

## Create installer custom resource

`installation/scripts/create-cr.sh` script prepares `installer` Custom Resource from `installation/resources/installer-cr.yaml.tpl` template. In local installation scenario default `installer` Custom Resource will be used. `kyma-installer` already contains local `kyma` resources bundled thus `url` is ignored by `installer` component. 

>**NOTE:** For `installer` Custom Resource details refer to the `docs/kyma/docs/040-cr-installation.md` file.

## Kyma-Installer deployment

The `installation/scripts/installer.sh` script creates the default RBAC role, installs [Tiller] (https://docs.helm.sh/), and deploys the `kyma-installer` component.

>**NOTE:** For `kyma-installer` deployment details refer to the `installation/resources/installer.yaml` file.

The script applies the `installer` Custom Resource and marks it with the `action=install` label, which triggers the Kyma installation.

>**NOTE:** `kyma` installation runs in background. Execute `./installation/scripts/is-installed.sh` to follow the installation process.