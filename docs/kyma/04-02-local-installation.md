---
title: Install Kyma locally
type: Installation
---

This Installation guide shows you how to quickly deploy Kyma locally on the MacOS and Linux platforms. Kyma is installed locally using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/). The document provides prerequisites and instructions on how to install Kyma on your machine, as well as the troubleshooting tips.

## Prerequisites

- [Kyma CLI](https://github.com/kyma-project/cli) 
- [Docker](https://www.docker.com/get-started)
- [Minikube](https://github.com/kubernetes/minikube) 1.0.1
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.12.5
- [Helm](https://github.com/kubernetes/helm) 2.10.0
- [jq](https://stedolan.github.io/jq/)
- [wget](https://www.gnu.org/software/wget/)

Virtualization:

- [Hyperkit driver](https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#hyperkit-driver) - MacOS only
- [VirtualBox](https://www.virtualbox.org/) - Linux only

> **NOTE:** To work with Kyma, use only the provided scripts and commands. Kyma requires a specific Minikube configuration and does not work on a basic Minikube cluster that you can start using the `minikube start` command.

## Install Kyma and access the Console UI

Follow there instructions to install Kyma from a release or from sources:
<div tabs>
  <details>
  <summary>
  From a release
  </summary>
    
  1. Provision a kubernetes cluster on minikube ready to be installed with Kyma run this command:
    
     ```bash
     kyma provision minikube
     ```
     > **NOTE:** Kyma CLI command above uses default minikube vm driver installed for your OS, for list of default vm driver for supported OS and other available vm drivers you can read from [here](http://github.com/kyma-project/cli)
  2. Install Kyma with latest release, run following command:
     ```bash
     kyma install
     ```
     > **NOTE** If you need install a specific Kyma release use following steps
     
  3. Go to [this](https://github.com/kyma-project/kyma/releases/) page and choose the latest release.  
  
  4. Export the release version as an environment variable:
     ```bash
     export KYMA_RELEASE={KYMA_RELEASE_VERSION}
     ```
  5. Install Kyma release `$KYMA_RELEASE` in minikube:
     ```bash
     kyma install -r $KYMA_RELEASE
     ```
     
     Wait until installation process completed
     
  6. After the installation is completed, you can access the Console UI. Go to [this](https://console.kyma.local) address and select **Login with Email**. Use the **admin@kyma.cx** email address and the password printed in the terminal after installation finished.
  
  7. At this point, Kyma is ready for you to explore. See what you can achieve using the Console UI or check out one of the [available examples](https://github.com/kyma-project/examples).
   
  </details>
  <details>
  <summary>
  From sources
  </summary>
    
  1. Open a terminal window and navigate to a space in which you want to store local Kyma sources.
    
  2. Clone the Kyma repository to your machine using either HTTPS or SHH. Run this command to clone the repository.
     
     <div tabs>
       <details>
       <summary>
       HTTPS
       </summary>
     
        ```bash
        git clone https://github.com/kyma-project/kyma.git
        ```
       </details>
       <details>
       <summary>
       SSH
       </summary>
     
       ```bash
       git clone git@github.com:kyma-project/kyma.git
       ```
       </details>
     </div>
 
  3. Kyma comes with a local wildcard self-signed `server.crt` certificate. Add this certificate to your OS trusted certificates to access the Console UI. On MacOS, run:
     ```bash
     sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain certs/workspace/raw/server.crt
     ```
     >**NOTE:** Mozilla Firefox uses its own certificate keychain. If you want to access the Console UI though Firefox, add the Kyma wildcard certificate to the certificate keychain of the browser. To access the Application Connector and connect an external solution to the local deployment of Kyma, you must add the certificate to the trusted certificate storage of your programming environment. Read [this](/components/application-connector#details-access-the-application-connector-on-a-local-kyma-deployment) document to learn more.

  4. Provision a kubernetes cluster on minikube ready to be installed with Kyma run this command:
     ```bash
     kyma provision minikube
     ```
     > **NOTE:** Kyma CLI command above uses default minikube vm driver installed for your OS, for list of default vm driver for supported OS and other available vm drivers you can read from [here](http://github.com/kyma-project/cli)
  5. Install Kyma from sources run following command:
     
     ```bash
     kyma install --local --src-path {YOUR_KYMA_SOURCE_PATH}
     ```
     
     Wait until installation process completed.
  
  6. After the installation is completed, you can access the Console UI. Go to [this](https://console.kyma.local) address and select **Login with Email**. Use the **admin@kyma.cx** email address and the password printed in the terminal after installation finished.

  7. At this point, Kyma is ready for you to explore. See what you can achieve using the Console UI or check out one of the [available examples](https://github.com/kyma-project/examples).

   </details>
</div>


Read [this](#installation-reinstall-kyma) document to learn how to reinstall Kyma without deleting the cluster from Minikube.
To learn how to test Kyma, see [this](#details-testing-kyma) document.

## Stop and restart Kyma without reinstalling

Use the `minikube.sh` script to restart the Minikube cluster without reinstalling Kyma. Follow these steps to stop and restart your cluster:

1. Stop the Minikube cluster with Kyma installed. Run:
   ```
   minikube stop
   ```
2. Restart the cluster without reinstalling Kyma. Run:
   ```bash
   kyma provision minikube
   ```

The Kyma CLI discovers that a minikube cluster is initialized and asks if you want to delete it. Answering `no` causes the Kyma CLI to start the minikube cluster and restarts all of the previously installed components. Even though this procedure takes some time, it is faster than a clean installation as you don't download all of the required Docker images.

To verify that the restart is successful, run this command and check if all Pods have the `RUNNING` status:

```
kubectl get pods --all-namespaces
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
