---
title: Local Kyma installation
type: Getting Started
---

## Overview

This Getting Started guide instructs developers to quickly deploy Kyma locally on a Mac, Linux, or Windows. Kyma installs locally using a proprietary installer based on a Kubernetes operator.
The document provides the prerequisites, the instructions on how to install Kyma locally and verify the deployment, and the troubleshooting tips.

## Prerequisites

To run Kyma locally, download these tools:

- [Minikube](https://github.com/kubernetes/minikube) 0.28.2
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.10.0
- [Helm](https://github.com/kubernetes/helm) 2.8.2
- [Hyperkit driver](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver) - Mac only
- [Virtualbox](https://www.virtualbox.org/) - Linux or Windows
- Hyper-V - Windows

Read the [prerequisite reasoning](019-prereq-reasoning.md) document to learn why Kyma uses these tools.

## Setup certificates

Kyma comes with a local wildcard self-signed [certificate](../../../installation/certs/workspace/raw/server.crt). Trust it on the OS level for convenience. Alternatively, accept exceptions for each subdomain in your browser as you use Kyma.

Follow these steps to "always trust" the Kyma certificate on macOS:

1. Open the Keychain Access application. Select **System** from the **Keychains** menu.
2. Go to **File**, select **Import items...**, and import the Kyma [certificate](../../../installation/certs/workspace/raw/server.crt).
3. Go to the **Certificates** view and find the `*.kyma.local` certificate you imported.
4. Right-click the certificate and select **Get Info**.
5. Expand the **Trust** list and set **When using this certificate** to **Always trust**.
6. Close the certificate information window and enter your system password to confirm the changes.

>**NOTE:**
- The process is complete when you close the certificate information window and enter your password. You don't get the expected results if you try to use the certificate before completing this step.
- "Always trusting" the certificate does not work with Mozilla Firefox.

## Install Kyma on Minikube

> **NOTE:** Running the installation script deletes any previously existing cluster from your Minikube.

1. Change the working directory to `installation`:
  ```bash
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

Read the [Reinstall Kyma](025-details-local-reinstallation.md) document to learn how to reinstall Kyma without deleting the cluster from Minikube.
To learn how to test Kyma, see the [Testing Kyma](026-details-testing.md) document.

### Custom Resource file

The Custom Resource file contains controls the Kyma installer, which is a proprietary solution based on the [Kubernetes operator](https://coreos.com/operators/). The file contains the basic information that defines Kyma installation.
Find the custom resource template [here](../../../installation/resources/installer-cr.yaml.tpl).

### Control the installation process

To trigger the installation process, set the **action** label to `install` in the metadata of the Custom Resource with the installer configuration.
To trigger the deinstallation process, set the **action** label to `uninstall` in the metadata of the Custom Resource with the installer configuration.

### Generate a new Custom Resource file

Use the `create-cr.sh` script to generate the Custom Resource file. The script accepts these arguments:

- `--output` - mandatory. The location of the Custom Resource output file
- `--url` - the URL of the Kyma package to install
- `--version` - the Kyma version
- `--ipaddr` - the load balancer IP
- `--domain` - the instance domain

For example:
```
$ ./installation/scripts/create-cr.sh --output kyma-cr.yaml --url {Kyma_TAR.GZ_URL} --version 0.0.1 --domain kyma.local
```

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

* Click **Login with Email** and sign in with the `admin@kyma.cx` email address and the generic password from the [Dex ConfigMap](../../../resources/dex/templates/dex-config-map.yaml) file.

* Click the **Environments** section and select an Environment from the drop-down menu to explore Kyma further.

### Access the Kubernetes Dashboard

Additionally, confirm that you can access your Kubernetes Dashboard. Run the following command to check the IP address on which Minikube is running:

```bash
minikube ip
```

The URL of your Kubernetes Dashboard looks similar to this:
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
