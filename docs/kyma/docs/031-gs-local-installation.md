---
title: Local Kyma installation
type: Getting Started
---

This Getting Started guide shows developers how to quickly deploy Kyma locally on a Mac, Linux, or Windows. Kyma installs locally using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/). The document provides prerequisites, instructions on how to install Kyma locally and verify the deployment, as well as the troubleshooting tips.

## Prerequisites

To run Kyma locally, clone this Git repository to your machine and checkout the `latest` tag. After you clone the repository, run this command:
```
git checkout latest
```
Additionally, download these tools:

- [Minikube](https://github.com/kubernetes/minikube) 0.28.2
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.10.0
- [Helm](https://github.com/kubernetes/helm) 2.8.2
- [jq](https://stedolan.github.io/jq/)

Virtualization:

- [Hyperkit driver](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver) - Mac only
- [VirtualBox](https://www.virtualbox.org/) - Linux or Windows
- [Hyper-V](https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/quick-start/enable-hyper-v) - Windows

> **NOTE:** To work with Kyma, use only the provided installation and deinstallation scripts. Kyma does not work on a basic Minikube cluster that you can start using the `minikube start` command or stop with the `minikube stop` command. If you don't need Kyma on Minikube anymore, remove the cluster with the `minikube delete` command.

## Set up certificates

Kyma comes with a local wildcard self-signed `server.crt` certificate that you can find under the `/installation/certs/workspace/raw/` directory of the `kyma` repository. Trust it on the OS level for convenience.

Follow these steps to "always trust" the Kyma certificate on Mac:

1. Change the working directory to `installation`:
  ```
  cd installation
  ```
2. Run this command:
  ```
  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain certs/workspace/raw/server.crt
  ```

>**NOTE:** "Always trusting" the certificate does not work with Mozilla Firefox.

## Install Kyma on Minikube

You can install Kyma with all core subcomponents or only with the selected ones. This section describes how to install all core subcomponents. To learn how to install only the specific ones, see the **Install subcomponents** document for details.

> **NOTE:** Running the installation script deletes any previously existing cluster from your Minikube.

1. Change the working directory to `installation`:
  ```
  cd installation
  ```

2. Depending on your operating system, run `run.sh` for Mac and Linux or `run.ps1` for Windows
  ```
  cmd/run.sh
  ```

The `run.sh` script does not show the progress of the Kyma installation, which allows you to perform other tasks in the terminal window. However, to see the status of the Kyma installation, run this script after you set up the cluster and the installer:

```
scripts/is-installed.sh
```

Read the **Reinstall Kyma** document to learn how to reinstall Kyma without deleting the cluster from Minikube.
To learn how to test Kyma, see the **Testing Kyma** document.

## Verify the deployment

Follow the guidelines in the subsections to confirm that your Kubernetes API Server is up and running as expected.

### Access Kyma with CLI

Verify the cluster deployment with the kubectl command line interface (CLI).

Run this command to fetch all Pods in all Namespaces:

  ``` bash
  kubectl get pods --all-namespaces
  ```
The command retrieves all Pods from all Namespaces, the status of the Pods, and their instance numbers. Check if the **STATUS** column shows `Running` for all Pods. If any of the Pods that you require do not start successfully, perform the installation again.

### Access the Kyma console

Access your local Kyma instance through [this](https://console.kyma.local/) link.

* Click **Login with Email** and sign in with the `admin@kyma.cx` email address and the generic password from the `dex-config-map.yaml` file in the `/resources/dex/templates/` directory.

* Click the **Environments** section and select an Environment from the drop-down menu to explore Kyma further.

### Access the Kubernetes Dashboard

Additionally, confirm that you can access your Kubernetes Dashboard. Run the following command to check the IP address on which Minikube is running:

```bash
minikube ip
```

The address of your Kubernetes Dashboard looks similar to this:
```
http://{ip-address}:30000
```

See the example of the website address:

```
http://192.168.64.44:30000
```

## Troubleshooting

If the installer does not respond as expected, check the installation status using the `is-installed.sh` script with the `--verbose` flag added. Run:
```
scripts/is-installed.sh --verbose
```