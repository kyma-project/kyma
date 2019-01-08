---
title: Local installation scripts
type: Installation
---

This document extends the **Install Kyma locally from sources** guide with a detailed breakdown of the alternative local installation method which is the `run.sh` script.

The following snippet is the main element of the `run.sh` script:

```
if [[ ! $SKIP_MINIKUBE_START ]]; then
    bash $CURRENT_DIR/../scripts/minikube.sh --domain "$DOMAIN" --vm-driver "$VM_DRIVER"
fi

bash $CURRENT_DIR/../scripts/build-kyma-installer.sh --vm-driver "$VM_DRIVER"

if [ -z "$CR_PATH" ]; then

    TMPDIR=`mktemp -d "$CURRENT_DIR/../../temp-XXXXXXXXXX"`
    CR_PATH="$TMPDIR/installer-cr-local.yaml"

    bash $CURRENT_DIR/../scripts/create-cr.sh --output "$CR_PATH" --domain "$DOMAIN"
    bash $CURRENT_DIR/../scripts/installer.sh --local --cr "$CR_PATH"

    rm -rf $TMPDIR
else
    bash $CURRENT_DIR/../scripts/installer.sh --cr "$CR_PATH"
fi
```
Subsequent sections provide details of all involved subscripts, in the order in which the `run.sh` script triggers them.

## The minikube.sh script

> **NOTE:** To work with Kyma, use only the provided scripts and commands. Kyma does not work on a basic Minikube cluster that you can start using the `minikube start` command or stop with the `minikube stop` command. If you don't need Kyma on Minikube anymore, remove the cluster with the `minikube delete` command.

The purpose of the `installation/scripts/minikube.sh` script is to configure and start Minikube. The script also checks if your development environment is configured to handle the Kyma installation. This includes checking Minikube and kubectl versions. If Minikube is already initialized, the system prompts you to agree to remove the previous Minikube cluster. The script exits if you do not want to restart your cluster.

Minikube is configured to disable the default Nginx Ingress Controller.

>**NOTE:** For the complete list of parameters passed to the `minikube start` command, refer to the `installation/scripts/minikube.sh` script.

Once Minikube is up and running, the script adds local installation entries to `/etc/hosts`.

## The build-kyma-installer.sh script

The Installer is an application based on a [Kubernetes operator](https://coreos.com/operators/). Its purpose is to install Helm charts defined in the Installation custom resource. The Kyma-Installer is a Docker image that bundles the Installer binary with Kyma charts.

The `installation/scripts/build-kyma-installer.sh` script extracts the Kyma-Installer image name from the `installer.yaml` deployment file and uses it to build a Docker image inside Minikube. This image contains local Kyma sources from the `resources` folder.

>**NOTE:** For the Kyma-Installer Docker image details, refer to the `tools/kyma-installer/kyma.Dockerfile` file.

## The configure-azure-broker.sh script

The `configure-azure-broker.sh` script configures Azure Broker, an optional subcomponent of the `core` deployment.

The Azure Broker subcomponent is part of the `core` deployment that provisions managed services in the Microsoft Azure cloud. To enable the Azure Broker, export the following environment variables:
 - AZURE_BROKER_SUBSCRIPTION_ID
 - AZURE_BROKER_TENANT_ID
 - AZURE_BROKER_CLIENT_ID
 - AZURE_BROKER_CLIENT_SECRET

>**NOTE:** You need to export above environment variables before executing the `installation/cmd/run.sh` script. As the Azure credentials are converted to a Kubernetes Secret, make sure the exported values are base64-encoded.

## The create-cr.sh script

The `installation/scripts/create-cr.sh` script prepares the Installation custom resource from the `installation/resources/installer-cr.yaml.tpl` template. The local installation scenario uses the default Installation custom resource. The Kyma-Installer already contains local Kyma resources bundled, thus `url` is ignored by the Installer component.

>**NOTE:** For the Installation custom resource details, refer to the **Installation** document.

## The installer.sh script

The `installation/scripts/installer.sh` script creates the default RBAC role, installs [Tiller](https://docs.helm.sh/), and deploys the Kyma-Installer component.

>**NOTE:** For the Kyma Installer deployment details, refer to the `installation/resources/installer.yaml` file.

The script applies the Installation custom resource and marks it with the `action=install` label, which triggers the Kyma installation.

>**NOTE:** The Kyma installation runs in the background. Execute the `./installation/scripts/is-installed.sh` script to follow the installation process.

## The is-installed.sh script

The `installation/scripts/is-installed.sh` script shows the status of Kyma installation in real time. The script checks the status of the Installation custom resource. When it detects that the status changed to `Installed`, the script exits. If you define a timeout period and the status doesn't change to `Installed` within that period, the script fetches the installer logs. If you don't set a timeout period, the script waits for the change of the status until you terminate it. 
