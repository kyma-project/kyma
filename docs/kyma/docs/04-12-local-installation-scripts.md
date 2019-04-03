---
title: Local installation scripts deep-dive
type: Installation
---

This document extends the [Install Kyma locally from sources](#installation-install-kyma-locally-from-sources) guide with a detailed breakdown of the alternative local installation method which is the `run.sh` script.

The following snippet is the main element of the `run.sh` script:

```
if [[ ! $SKIP_MINIKUBE_START ]]; then
    bash $SCRIPTS_DIR/minikube.sh --domain "$DOMAIN" --vm-driver "$VM_DRIVER" $MINIKUBE_EXTRA_ARGS
fi

bash $SCRIPTS_DIR/build-kyma-installer.sh --vm-driver "$VM_DRIVER"

if [ -z "$CR_PATH" ]; then

    TMPDIR=`mktemp -d "$CURRENT_DIR/../../temp-XXXXXXXXXX"`
    CR_PATH="$TMPDIR/installer-cr-local.yaml"
    bash $SCRIPTS_DIR/create-cr.sh --output "$CR_PATH" --domain "$DOMAIN"

fi

bash $SCRIPTS_DIR/installer.sh --local --cr "$CR_PATH" --password "$ADMIN_PASSWORD"
rm -rf $TMPDIR
```
Subsequent sections provide details of all involved subscripts, in the order in which the `run.sh` script triggers them.

## The minikube.sh script

> **NOTE:** To work with Kyma, use only the provided scripts and commands. Kyma does not work on a basic Minikube cluster that you can start using the `minikube start` command.

The purpose of the `installation/scripts/minikube.sh` script is to configure and start Minikube. The script also checks if your development environment is configured to handle the Kyma installation. This includes checking Minikube and kubectl versions.

If Minikube is already initialized, the system prompts you to agree to remove the previous Minikube cluster.
- If you plan to perform a clean installation, answer `yes`.
- If you installed Kyma to your Minikube cluster and then stopped the cluster using the `minikube stop` command, answer `no`.  This allows you to start the cluster again without reinstalling Kyma.

Minikube is configured to disable the default Nginx Ingress Controller.

>**NOTE:** For the complete list of parameters passed to the `minikube start` command, refer to the `installation/scripts/minikube.sh` script.

Once Minikube is up and running, the script adds local installation entries to `/etc/hosts`.

## The build-kyma-installer.sh script

The Kyma Installer is an application based on the [Kubernetes operator](https://coreos.com/operators/). Its purpose is to install Helm charts defined in the Installation custom resource. The Kyma Installer is a Docker image that bundles the Installer binary with Kyma charts.

The `installation/scripts/build-kyma-installer.sh` script extracts the Kyma-Installer image name from the `installer.yaml` deployment file and uses it to build a Docker image inside Minikube. This image contains local Kyma sources from the `resources` folder.

>**NOTE:** For the Kyma Installer Docker image details, refer to the `tools/kyma-installer/kyma.Dockerfile` file.

## The create-cr.sh script

The `installation/scripts/create-cr.sh` script prepares the Installation custom resource from the `installation/resources/installer-cr.yaml.tpl` template. The local installation scenario uses the default Installation custom resource. The Kyma Installer already contains local Kyma resources bundled, thus `url` is ignored by the Installer component.

>**NOTE:** Read [this](#custom-resource-installation) document to learn more about the Installation custom resource.

## The installer.sh script

The `installation/scripts/installer.sh` script creates the default RBAC role, installs [Tiller](https://docs.helm.sh/), and deploys the Kyma Installer component.

>**NOTE:** For the Kyma Installer deployment details, refer to the `installation/resources/installer.yaml` file.

The script applies the Installation custom resource and marks it with the `action=install` label, which triggers the Kyma installation.

In the process of installing Tiller, a set of TLS certificates is created and saved to [Helm Home](https://helm.sh/docs/glossary/#helm-home-helm-home) to secure the connection between the client and the server.

>**NOTE:** The Kyma installation runs in the background. Execute the `./installation/scripts/is-installed.sh` script to follow the installation process.

## The is-installed.sh script

The `installation/scripts/is-installed.sh` script shows the status of Kyma installation in real time. The script checks the status of the Installation custom resource. When it detects that the status changed to `Installed`, the script exits. If you define a timeout period and the status doesn't change to `Installed` within that period, the script fetches the Installer logs. If you don't set a timeout period, the script waits for the change of the status until you terminate it.
