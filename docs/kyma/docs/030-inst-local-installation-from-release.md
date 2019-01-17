---
title: Install Kyma locally from the release
type: Installation
---

This Installation guide shows developers how to quickly deploy Kyma locally on a Mac or Linux from the latest release. Kyma installs locally using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/). The document provides prerequisites, instructions on how to install Kyma locally and verify the deployment, as well as the troubleshooting tips.

## Prerequisites

To run Kyma locally, clone this Git repository to your machine and check out the latest release.

Additionally, download these tools:

- [Minikube](https://github.com/kubernetes/minikube) 0.28.2
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.10.0
- [Helm](https://github.com/kubernetes/helm) 2.10.0
- [jq](https://stedolan.github.io/jq/)

Virtualization:

- [Hyperkit driver](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver) - Mac only
- [VirtualBox](https://www.virtualbox.org/) - Linux only

> **NOTE:** To work with Kyma, use only the provided scripts and commands. Kyma does not work on a basic Minikube cluster that you can start using the `minikube start` command.

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

You can install Kyma either with all core subcomponents or only with the selected ones. This section describes how to install Kyma with all core subcomponents. To learn how to install only the specific ones, see the **Install subcomponents** document for details.

> **NOTE:** Running the installation script deletes any previously existing cluster from your Minikube.

> **NOTE:** Logging and Monitoring subcomponents are not included by default when you install Kyma on Minikube. You can install them using the instructions provided [here](../../../resources/README.md).

To install Kyma, follow these steps:

1. Change the working directory to `installation`:
  ```
  cd installation
  ```

2. Use the following command to run Kubernetes locally using Minikube:
  ```
  ./scripts/minikube.sh --domain "kyma.local" --vm-driver "hyperkit"
  ```

3. Wait until the `kube-dns` Pod is ready. Run this script to setup Tiller:
  ```
  ./scripts/install-tiller.sh
  ```

4. Go to [this](https://github.com/kyma-project/kyma/releases/) page and choose the release you want to use.

5. Export the version you chose as an environment variable. Run:
  ```
  export LATEST={KYMA_RELEASE_VERSION}
  ```

6. Configure the Kyma installation using the local configuration file from the $LATEST release:
  ```
  kubectl apply -f https://github.com/kyma-project/kyma/releases/download/$LATEST/kyma-config-local.yaml
  ```

7. To trigger the installation process, label the `kyma-installation` custom resource:
  ```
  kubectl label installation/kyma-installation action=install
  ```

8. By default, the Kyma installation is a background process, which allows you to perform other tasks in the terminal window. Nevertheless, you can track the progress of the installation by running this script:
  ```
  ./scripts/is-installed.sh
  ```

Read the **Reinstall Kyma** document to learn how to reinstall Kyma without deleting the cluster from Minikube.
To learn how to test Kyma, see the **Testing Kyma** document.

## Verify the deployment

Follow the guidelines in the subsections to confirm that your Kubernetes API Server is up and running as expected.

### Verify the installation status using the is-installed.sh script

The `is-installed.sh` script is designed to give you clear information about the Kyma installation. Run it at any point to get the current installation status, or to find out whether the installation is successful.

If the script indicates that the installation failed, try to install Kyma again by re-running the `run.sh` script.

If the installation fails in a reproducible manner, don't hesitate to create a [GitHub](https://github.com/kyma-project/kyma/issues) issue in the project or reach out to the ["installation" Slack channel](https://kyma-community.slack.com/messages/CD2HJ0E78) to get direct support from the community.

### Access the Kyma console

Access your local Kyma instance through [this](https://console.kyma.local/) link.

* Click **Login with Email** and sign in with the **admin@kyma.cx** email address. Use the password contained in the  `admin-user` Secret located in the `kyma-system` Namespace. To get the password, run:

``` bash
kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 -D
```

* Click the **Namespaces** section and select a Namespace from the drop-down menu to explore Kyma further.

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

## Enable Horizontal Pod Autoscaler (HPA)

By default, the Horizontal Pod Autoscaler is not enabled in your local Kyma installation, so you need to enable it manually.

Kyma uses the autoscaling/v1 stable version, which only provides support for CPU autoscaling. Once enabled, HPA automatically scales the number of lambda function Pods based on observed CPU utilization.

>**NOTE:** The autoscaling/v1 version does not support custom metrics. To use such metrics, you need the autoscaling/v2beta2 version.

To enable Horizontal Pod Autoscaler, follow these steps:

1. Enable the metrics server for resource metrics by running the following command:
  ```
  minikube addons enable metrics-server
  ```

2. Verify if the metrics server is active by checking the list of addons:
  ```
  minikube addons list
  ```

## Stop and restart Kyma without reinstalling

Use the `minikube.sh` script to restart the Minikube cluster without reinstalling Kyma. Follow these steps to stop and restart your cluster:

1. Stop the Minikube cluster with Kyma installed. Run:
```
minikube stop
```
2. Restart the cluster without reinstalling Kyma. Run:
```
./scripts/minikube.sh --domain "kyma.local" --vm-driver "hyperkit"
```

The script discovers that a minikube cluster is initialized and asks if you want to delete it. Answering `no` causes the script to start the Minikube cluster and restarts all of the previously installed components. Even though this procedure takes some time, it is faster than a clean installation as you don't download all of the required Docker images.

To verify that the restart is successful, run this command and check if all Pods have the `RUNNING` status:

```
kubectl get pods --all-namespaces
```

## Troubleshooting

If the Installer does not respond as expected, check the installation status using the `is-installed.sh` script with the `--verbose` flag added. Run:
```
scripts/is-installed.sh --verbose
```

If the installation is successful but a component does not behave in an expected way, see if all deployed Pods are running. Run this command:  
```
kubectl get pods --all-namespaces
```

The command retrieves all Pods from all Namespaces, the status of the Pods, and their instance numbers. Check if the STATUS column shows Running for all Pods. If any of the Pods that you require do not start successfully, perform the installation again.
If the problem persists, don't hesitate to create a [GitHub](https://github.com/kyma-project/kyma/issues) issue or reach out to the ["installation" Slack channel](https://kyma-community.slack.com/messages/CD2HJ0E78) to get direct support from the community.
