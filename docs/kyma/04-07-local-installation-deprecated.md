---
title: Install Kyma locally (Scripts)
type: Installation
---

>**WARNING:** Will be deprecated {TBD} with Kyma release 1.14

This Installation guide shows you how to quickly deploy Kyma locally on the MacOS and Linux platforms. Kyma is installed locally using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/). The document provides prerequisites and instructions on how to install Kyma on your machine, as well as the troubleshooting tips.

## Prerequisites

- [Docker](https://www.docker.com/get-started)
- [Minikube](https://github.com/kubernetes/minikube) 0.33.0
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.12.0
- [Helm](https://github.com/kubernetes/helm) 2.10.0
- [jq](https://stedolan.github.io/jq/)
- [wget](https://www.gnu.org/software/wget/)

Virtualization:

- [Hyperkit driver](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver) - MacOS only
- [VirtualBox](https://www.virtualbox.org/) - Linux only

> **NOTE:** To work with Kyma, use only the provided scripts and commands. Kyma requires a specific Minikube configuration and does not work on a basic Minikube cluster that you can start using the `minikube start` command.

## Install Kyma and access the Console UI

1. Open a terminal window and navigate to a space in which you want to store your local Kyma sources.

2. Clone the Kyma repository to your machine using either HTTPS or SSH. Run this command to clone the repository and change your working directory to `kyma`:
    <div tabs>
      <details>
      <summary>
      HTTPS
      </summary>

      ```
      git clone https://github.com/kyma-project/kyma.git ; cd kyma
      ```
      </details>
      <details>
      <summary>
      SSH
      </summary>

      ```
      git clone git@github.com:kyma-project/kyma.git ; cd kyma
      ```
      </details>
    </div>

3. Choose the release from which you want to install Kyma. Go to the [GitHub releases page](https://github.com/kyma-project/kyma/releases) to find out more about each of the available releases. Run this command to list all of the available tags that correspond to releases:
  ```
  git tag
  ```

4. Checkout the tag corresponding to the Kyma release you want to install. Run:
  ```
  git checkout {TAG}
  ```
 >**NOTE:** If you don't checkout any of the available tags, your sources match the state of the project's `master` branch at the moment of cloning the repository.

5. Navigate to the `installation` directory which contains all of the required installation resources. Run:
  ```
  cd installation
  ```

6. Kyma comes with a local wildcard self-signed `server.crt` certificate. Add this certificate to your OS trusted certificates to access the Console UI. On MacOS, run:
  ```
  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain certs/workspace/raw/server.crt
  ```
  >**NOTE:** Mozilla Firefox uses its own certificate keychain. If you want to access the Console UI though Firefox, add the Kyma wildcard certificate to the certificate keychain of the browser. To access the Application Connector and connect an external solution to the local deployment of Kyma, you must add the certificate to the trusted certificate storage of your programming environment. Read [this](/components/application-connector#details-access-the-application-connector-on-a-local-kyma-deployment) document to learn more.

7. Start the installation. Trigger the `run.sh` script to start Minikube with a Kyma-specific configuration and install the necessary components. Define the password used to log in to the Console UI using the `--password` flag. Run:
    <div tabs>
      <details>
      <summary>
      MacOS
      </summary>

      ```
      ./cmd/run.sh --password {USER_PASSWORD}
      ```
      </details>
      <details>
      <summary>
      Linux
      </summary>

      ```
      ./cmd/run.sh --password {USER_PASSWORD} --vm-driver virtualbox
      ```
      </details>
    </div>

8. By default, the Kyma installation is a background process, which allows you to perform other tasks in the terminal window. Nevertheless, you can track the progress of the installation by running the `is-installed.sh` script. It is designed to give you clear information about the Kyma installation. Run it at any point to get the current installation status, or to find out whether the installation is successful.

  ```
  ./scripts/is-installed.sh
  ```
  >**TIP:** If the script indicates that the installation failed, try to install Kyma again by re-running the `run.sh` script. If the installation fails in a reproducible manner, don't hesitate to create a [GitHub](https://github.com/kyma-project/kyma/issues) issue in the project or reach out to the ["installation" Slack channel](https://kyma-community.slack.com/messages/CD2HJ0E78) to get direct support from the community.

9. After the installation is completed, you can access the Console UI. Go to [this](https://console.kyma.local) address and select **Login with Email**. Use the **admin@kyma.cx** email address and the password you set using the `--password` flag.

10. At this point, Kyma is ready for you to explore. See what you can achieve using the Console UI or check out one of the [available examples](https://github.com/kyma-project/examples).

Read [this](#installation-reinstall-kyma) document to learn how to reinstall Kyma without deleting the cluster from Minikube.
To learn how to test Kyma, see [this](#details-testing-kyma) document.

## Stop and restart Kyma without reinstalling

Use the `minikube.sh` script to restart the Minikube cluster without reinstalling Kyma. Follow these steps to stop and restart your cluster:

1. Stop the Minikube cluster with Kyma installed. Run:
  ```
  minikube stop
  ```
2. Restart the cluster without reinstalling Kyma. Run:
    <div tabs>
      <details>
      <summary>
      MacOS
      </summary>

      ```
      ./scripts/minikube.sh --domain "kyma.local" --vm-driver hyperkit
      ```
      </details>
      <details>
      <summary>
      Linux
      </summary>

      ```
      ./scripts/minikube.sh --domain "kyma.local" --vm-driver virtualbox
      ```
      </details>
    </div>

The script discovers that a Minikube cluster is initialized and asks if you want to delete it. Answering `no` causes the script to start the Minikube cluster and restarts all of the previously installed components. Even though this procedure takes some time, it is faster than a clean installation as you don't download all of the required Docker images.

To verify that the restart is successful, run this command and check if all Pods have the `RUNNING` status:

```
kubectl get pods --all-namespaces
```

## Enable Horizontal Pod Autoscaler (HPA)

By default, the Horizontal Pod Autoscaler (HPA) is not enabled in your local Kyma installation, so you need to enable it manually.

Kyma uses the `autoscaling/v1` stable version, which supports only CPU autoscaling. Once enabled, HPA automatically scales the number of lambda function Pods based on the observed CPU utilization.

>**NOTE:** The `autoscaling/v1` version does not support custom metrics. To use such metrics, you need the `autoscaling/v2beta1` version.

Follow these steps to enable HPA:

1. Enable the metrics server for resource metrics by running the following command:
  ```
  minikube addons enable metrics-server
  ```

2. Verify if the metrics server is active by checking the list of addons:
  ```
  minikube addons list
  ```

## Troubleshooting

1. If you don't set the password for the **admin@kyma.cx** user using the `--password` parameter or you forget the password you set, you can get it from the `admin-user` Secret located in the `kyma-system` Namespace. Run this command:
    ```
    kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode
    ```

2. If the Installer does not respond as expected, check the installation status using the `is-installed.sh` script with the `--verbose` flag added. Run:
   ```
   scripts/is-installed.sh --verbose
   ```

3. If the installation is successful but a component does not behave in an expected way, see if all deployed Pods are running. Run this command:
   ```
   kubectl get pods --all-namespaces
   ```

   The command retrieves all Pods from all Namespaces, the status of the Pods, and their instance numbers. Check if the STATUS column shows Running for all Pods. If any of the Pods that you require do not start successfully, perform the installation again.

   If the problem persists, don't hesitate to create a [GitHub](https://github.com/kyma-project/kyma/issues) issue or reach out to the ["installation" Slack channel](https://kyma-community.slack.com/messages/CD2HJ0E78) to get direct support from the community.

4. If you put your local running cluster into hibernation or use `minikube stop` and `minikube start` the date and time settings of Minikube get out of sync with the system date and time settings. As a result, the access token used to log in cannot be properly validated by Dex and you cannot log in to the console. To fix that, set the date and time used by your machine in Minikube. Run:
   ```
   minikube ssh -- docker run -i --rm --privileged --pid=host debian nsenter -t 1 -m -u -n -i date -u $(date -u +%m%d%H%M%Y)
   ```
