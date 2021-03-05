---
title: Deploy Kyma locally
type: Installation
---

This Deployment guide shows you how to quickly deploy Kyma locally on the MacOS, Linux, and Windows platforms. Kyma is deployed locally using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/).

**NOTE:** Local deployment using Minikube is deprecated.

>**TIP:** See the [troubleshooting guide](#troubleshooting-overview) for tips.

## Prerequisites

<div tabs name="prerequisites" group="deploy-kyma-locally">
  <details>
  <summary label="k3s">
  k3s
  </summary>

- [Kyma CLI](https://github.com/kyma-project/cli)
- [Docker](https://www.docker.com/get-started)
- [k3s](https://k3s.io/) 

</details>
  <details>
  <summary label="minikube-deprecated">
  Minikube (deprecated)
  </summary>

- [Kyma CLI](https://github.com/kyma-project/cli)
- [Docker](https://www.docker.com/get-started)
- [Minikube](https://github.com/kubernetes/minikube) 1.12.2

</details>

Virtualization:

- [Hyperkit driver](https://minikube.sigs.k8s.io/docs/reference/drivers/hyperkit/) - MacOS only
- [VirtualBox](https://www.virtualbox.org/) - Linux only

> **NOTE**: To work with Kyma, use only the provided commands. Kyma requires a specific Minikube configuration and does not work on a basic Minikube cluster that you can start using the `minikube start` command.

## Deploy Kyma

Follow these instructions to deploy Kyma from a release or from sources:

<div tabs name="deploy-kyma" group="deploy-kyma-locally">
  <details>
  <summary label="from-a-release-k3s">
  From a release - k3s
  </summary>

  1. Provision a Kubernetes cluster on k3s. Run:

     ```bash
     kyma provision k3s
     ```

  2. Deploy the latest Kyma release on k3s:
     ```bash
     kyma deploy
     ```

     >**NOTE:** If you want to deploy a specific release version, go to the [GitHub releases page](https://github.com/kyma-project/kyma/releases) to find out more about available releases. Use the release version as a parameter when calling `kyma deploy --source {KYMA_RELEASE}`.

  </details>
  <details>
  <summary label="from-a-release-minikube">
  From a release - Minikube (deprecated)
  </summary>

  1. Provision a Kubernetes cluster on Minikube. Run:

     ```bash
     kyma provision minikube
     ```

  2. Deploy the latest Kyma release on Minikube:
     ```bash
     kyma deploy
     ```

     >**NOTE:** If you want to deploy a specific release version, go to the [GitHub releases page](https://github.com/kyma-project/kyma/releases) to find out more about available releases. Use the release version as a parameter when calling `kyma deploy --source {KYMA_RELEASE}`.

  </details>
  <details>
  <summary label="from-sources">
  From sources
  </summary>

  1. Open a terminal window and navigate to a space in which you want to store local Kyma sources.

  2. Clone the `Kyma` repository using HTTPS. Run:

     ```bash
     git clone https://github.com/kyma-project/kyma.git
     ```

  3. Provision a Kubernetes cluster on k3s. Run:

     ```bash
     kyma provision k3s
     ```

  4. Deploy Kyma from sources. Run:

     ```bash
     kyma deploy --source local
     ```
     >**NOTE:** By default, the deployment uses sources located under your [GOPATH](https://github.com/golang/go/wiki/GOPATH). If you want to use a specific source folder, use it as a parameter when calling `kyma deploy --source local --src-path {YOUR_KYMA_SOURCE_PATH}`.

   </details>
</div>

## Post-deployment steps

Kyma comes with a local wildcard self-signed `server.crt` certificate. The `kyma deploy` command downloads and adds this certificate to the trusted certificates in your OS so you can access the Console UI.

>**NOTE:** Mozilla Firefox uses its own certificate keychain. If you want to access the Console UI though Firefox, add the Kyma wildcard certificate to the certificate keychain of the browser. To access the Application Connector and connect an external solution to the local deployment of Kyma, you must add the certificate to the trusted certificate storage of your programming environment. See the [Java environment](/components/application-connector#details-access-the-application-connector-on-a-local-kyma-deployment) as an example.

1. After the deployment is completed, you can access the Console UI. To open the the Console UI in your default browser, run:

   ```bash
   kyma console
   ```

2. Select **Login with Email**. Use the **admin@kyma.cx** email address and the password printed in the terminal once the deployment process is completed.

3. At this point, Kyma is ready for you to explore. See what you can achieve using the Console UI or check out one of the [available examples](https://github.com/kyma-project/examples).

Learn also how to [test Kyma](#details-testing-kyma) or [redeploy](#installation-reinstall-kyma) it without deleting the cluster from k3s.

## Stop and restart Kyma without redeploying

Use the Kyma CLI to restart the k3s cluster without redeploying Kyma. Follow these steps to stop and restart your cluster:

1. Stop the k3s cluster with Kyma deployed. Run:

   ```bash
   k3s stop
   ```

2. Restart the cluster without redeploying Kyma. Run:

   ```bash
   kyma provision k3s
   ```

The Kyma CLI discovers that a k3s cluster is initialized and asks if you want to delete it. Answering `no` causes the Kyma CLI to start the k3s cluster and restarts all of the previously deployed components. Even though this procedure takes some time, it is faster than a clean deployment as you don't download all of the required Docker images.
