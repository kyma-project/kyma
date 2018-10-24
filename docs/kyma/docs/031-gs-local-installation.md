---
title: Local Kyma installation
type: Getting Started
---

This Getting Started guide shows developers how to quickly deploy Kyma locally on a Mac or Linux. Kyma installs locally using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/). The document provides prerequisites, instructions on how to install Kyma locally and verify the deployment, as well as the troubleshooting tips.

## Prerequisites

To run Kyma locally, clone this Git repository to your machine and checkout the latest release.

Additionally, download these tools:

- [Minikube](https://github.com/kubernetes/minikube) 0.28.2
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.10.0
- [Helm](https://github.com/kubernetes/helm) 2.10.0
- [jq](https://stedolan.github.io/jq/)

Virtualization:

- [Hyperkit driver](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver) - Mac only
- [VirtualBox](https://www.virtualbox.org/) - Linux only

> **NOTE:** To work with Kyma, use only the provided scripts and commands. Kyma does not work on a basic Minikube cluster that you can start using the `minikube start` command or stop with the `minikube stop` command. If you don't need Kyma on Minikube anymore, remove the cluster with the `minikube delete` command.

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

To access the Application Connector and connect an external solution to the local deployment of Kyma, you must add the certificate to the trusted certificate storage of your programming environment. Read the **Access Application Connector on local Kyma** in the **Application Connector** topic to learn more.


## Install Kyma on Minikube

You can install Kyma with all core subcomponents or only with the selected ones. This section describes how to install Kyma from the latest release, with all core subcomponents. To learn how to install only the specific ones, see the **Install subcomponents** document for details.

> **NOTE:** Running the installation script deletes any previously existing cluster from your Minikube.

1. Change the working directory to `installation`:
  ```
  cd installation
  ```

2. Use the following command to run Kubernetes locally using Minikube:
```
$ ./scripts/minikube.sh --domain "kyma.local" --vm-driver "hyperkit"
```

3. Kyma installation requires increased permissions granted by the **cluster-admin** role. To bind the role to the default **ServiceAccount**, run the following command:
```
$ kubectl apply -f ./resources/default-sa-rbac-role.yaml
```

4. Wait until the `kube-dns` Pod is ready. Run this script to setup Tiller:
```
$ ./scripts/install-tiller.sh
```

5. Configure the Kyma installation using the local configuration file from the 0.4.3 release:
```
$ kubectl apply -f https://github.com/kyma-project/kyma/releases/download/0.4.3/kyma-config-local.yaml
```

6. To trigger the installation process, label the `kyma-installation` custom resource:
```
$ kubectl label installation/kyma-installation action=install
```

7. By default, the Kyma installation is a background process, which allows you to perform other tasks in the terminal window. Nevertheless, you can track the progress of the installation by running this script:
```
$ ./scripts/is-installed.sh
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

* Click **Login with Email** and sign in with the `admin@kyma.cx` email address. Use the password contained in the  `admin-user` Secret located in the `kyma-system` Namespace. To get the password, run:

``` bash
kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 -D
```

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
