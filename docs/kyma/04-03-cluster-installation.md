---
title: Install Kyma on a cluster
type: Installation
---

This installation guide explains how you can quickly deploy Kyma on a cluster with a wildcard DNS provided by [`xip.io`](http://xip.io) using a GitHub release of your choice.

>**NOTE:** If you want to expose the Kyma cluster on your own domain, follow [this](#installation-use-your-own-domain) installation guide. To install using your own image instead of a GitHub release, follow [these](#installation-use-your-own-kyma-installer-image) instructions.

If you need to use Helm and access Tiller, complete the [additional configuration](#installation-use-helm) after the installation.

Choose the installation type and get started:

<div tabs name="provider-installation">
  <details>
  <summary>
  GCP Marketplace
  </summary>

1. Access **project Kyma** on the [Google Cloud Platform (GCP) Marketplace](https://console.cloud.google.com/marketplace/details/sap-public/kyma?q=kyma%20project) and click **CONFIGURE**.

2. When the pop-up box appears, select you project from the available list and confirm your choice.

3. To create a Kubernetes cluster for your Kyma installation, select a cluster zone from the drop-down menu and click **Create cluster**. Wait for a few minutes for the Kubernetes cluster to deploy.

4. Leave the default values or adjust these settings:

  | Field   |      Default value     |
  |----------|-------------|
  | **Namespace** | `default` |
  | **App instance name** | `kyma-1` |
  | **Cluster Admin Service Account** | `Create a new service account` |

5. Accept the GCP Marketplace Terms of Service to continue.

6. Click the **Deploy** button for the Kyma installation to start.

> **NOTE:** The installation can take several minutes to complete.

7. Once you become redirected to the **Applications** page under **Kubernetes Engine** in the GCP Console, you get the installation status details. Check the installation status. If it is green, follow the steps under the **Next steps** section in **INFO PANEL** to import the self-signed TLS certificate to your trusted certified authorities.

8. Access the cluster using the link and login details provided in the **Kyma info** section on the **Application details** page.

> **TIP:** Watch [this](https://www.youtube.com/watch?v=hxVhQqI1B5A) video for a walkthrough of the installation process.

  </details>
  <details>
  <summary>
  GKE
  </summary>

Install Kyma on a [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/) (GKE) cluster.

## Prerequisites

- [Google Cloud Platform](https://console.cloud.google.com/) (GCP) project with Kubernetes Engine API enabled
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.12.0
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- [wget](https://www.gnu.org/software/wget/)


>**NOTE:** Running Kyma on GKE requires three [`n1-standard-4` machines](https://cloud.google.com/compute/docs/machine-types). You create these machines when you complete the **Prepare the GKE cluster** step.

## Choose the release to install

1. Go to [this](https://github.com/kyma-project/kyma/releases/) page and choose the release you want to install.

2. Export the release version as an environment variable. Run:

    ```
    export KYMA_VERSION={KYMA_RELEASE_VERSION}
    ```

## Prepare the GKE cluster

1. Select a name for your cluster. Export the cluster name, the name of your GCP project, and the zone you want to deploy to as environment variables. Run:

    ```
    export CLUSTER_NAME={CLUSTER_NAME_YOU_WANT}
    export GCP_PROJECT={YOUR_GCP_PROJECT}
    export GCP_ZONE={GCP_ZONE_TO_DEPLOY_TO}
    ```

2. Create a cluster in the zone defined in the previous step. Run:

    ```
    gcloud container --project "$GCP_PROJECT" clusters \
    create "$CLUSTER_NAME" --zone "$GCP_ZONE" \
    --cluster-version "1.12" --machine-type "n1-standard-4" \
    --addons HorizontalPodAutoscaling,HttpLoadBalancing
    ```

3. Configure kubectl to use your new cluster. Run:

    ```
    gcloud container clusters get-credentials $CLUSTER_NAME --zone $GCP_ZONE --project $GCP_PROJECT
    ```

4. Add your account as the cluster administrator:

    ```
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)
    ```

5. Install Tiller on your GKE cluster. Run:

    ```
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/$KYMA_VERSION/installation/resources/tiller.yaml
    ```

## Install Kyma

1. Deploy Kyma. Run:

    ```
    kubectl apply -f https://github.com/kyma-project/kyma/releases/download/$KYMA_VERSION/kyma-installer-cluster.yaml
    ```

2. Check if the Pods of Tiller and the Kyma Installer are running:

    ```
    kubectl get pods --all-namespaces
    ```

3. To watch the installation progress, run:

    ```
    while true; do \
      kubectl -n default get installation/kyma-installation -o jsonpath="{'Status: '}{.status.state}{', description: '}{.status.description}"; echo; \
      sleep 5; \
    done
    ```

After the installation process is finished, the `Status: Installed, description: Kyma installed` message appears.

If you receive an error, fetch the Kyma Installer logs using this command:

  ```
  kubectl -n kyma-installer logs -l 'name=kyma-installer'
  ```

## Post-installation steps

### Add the xip.io self-signed certificate to your OS trusted certificates

After the installation, add the custom Kyma [`xip.io`](http://xip.io/) self-signed certificate to the trusted certificates of your OS. For MacOS, run:

  ```
  tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
  && kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.tlsCrt}' | base64 --decode > $tmpfile \
  && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
  && rm $tmpfile
  ```

### Access the cluster

1. To get the address of the cluster's Console, check the host of the Console's virtual service. The name of the host of this virtual service corresponds to the Console URL. To get the virtual service host, run:

    ```
    kubectl get virtualservice core-console -n kyma-system -o jsonpath='{ .spec.hosts[0] }'
    ```

2. Access your cluster under this address:

    ```
    https://{VIRTUAL_SERVICE_HOST}
    ```

3. To log in to your cluster's Console UI, use the default `admin` static user. Click **Login with Email** and sign in with the **admin@kyma.cx** email address. Use the password contained in the `admin-user` Secret located in the `kyma-system` Namespace. To get the password, run:

    ```
    kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode
    ```

  </details>
  <details>
  <summary>
  AKS
  </summary>


Install Kyma on an [Azure Kubernetes Service](https://azure.microsoft.com/services/kubernetes-service/) (AKS) cluster.

## Prerequisites

- [Microsoft Azure](https://azure.microsoft.com)
- [Kubernetes](https://kubernetes.io/) 1.12 or higher
- Tiller 2.10.0 or higher
- [Docker](https://www.docker.com/)
- [Docker Hub](https://hub.docker.com/) account
- [az](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)


>**NOTE:** Running Kyma on AKS requires three [`Standard_D4_v3` machines](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/sizes-general#dv3-series-1). You create these machines when you complete the **Prepare the AKS cluster** step.

>**CAUTION:** Due to a known Istio-related issue, Kubernetes jobs run endlessly on AKS deployments of Kyma. Read [this](/components/service-mesh/#troubleshooting-kubernetes-jobs-fail-on-aks) document to learn more.

## Choose the release to install

1. Go to [this](https://github.com/kyma-project/kyma/releases/) page and choose the release you want to install.

2. Export the release version as an environment variable. Run:

    ```
    export KYMA_VERSION={KYMA_RELEASE_VERSION}
    ```

## Prepare the AKS cluster

1. Select a name for your cluster. Set the cluster name, the resource group and region as environment variables. Run:

    ```
    export RS_GROUP={YOUR_RESOURCE_GROUP_NAME}
    export CLUSTER_NAME={YOUR_CLUSTER_NAME}
    export REGION={YOUR_REGION} #westeurope
    ```

2. Create a resource group that will contain all your resources:

    ```
    az group create --name $RS_GROUP --location $REGION
    ```

3. Create an AKS cluster. Run:

    ```
    az aks create \
      --resource-group $RS_GROUP \
      --name $CLUSTER_NAME \
      --node-vm-size "Standard_D4_v3" \
      --kubernetes-version 1.12 \
      --enable-addons "monitoring,http_application_routing" \
      --generate-ssh-keys
    ```

4. To configure kubectl to use your new cluster, run:

    ```
    az aks get-credentials --resource-group $RS_GROUP --name $CLUSTER_NAME
    ```

5. Install Tiller and add additional privileges to be able to access readiness probes endpoints on your AKS cluster.

* Installation from release:

    ```
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/$KYMA_RELEASE_VERSION/installation/resources/tiller.yaml
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/$KYMA_RELEASE_VERSION/installation/resources/azure-crb-for-healthz.yaml
    ```

* If you install Kyma from sources, check out [kyma-project](https://github.com/kyma-project/kyma) and enter the root folder. Run:

    ```
    kubectl apply -f installation/resources/tiller.yaml
    kubectl apply -f installation/resources/azure-crb-for-healthz.yaml
    ```

6. Install custom installation overrides for AKS. Run:

    ```
    kubectl create namespace kyma-installer \
    && kubectl create configmap aks-overrides -n kyma-installer --from-literal=global.proxy.excludeIPRanges=10.0.0.1 \
    && kubectl label configmap aks-overrides -n kyma-installer installer=overrides component=istio
    ```

    >**TIP:** An example config map is available [here](./assets/aks-overrides.yaml).

## Install Kyma

1. Deploy Kyma. Run:

    ```
    kubectl apply -f https://github.com/kyma-project/kyma/releases/download/$KYMA_VERSION/kyma-installer-cluster.yaml
    ```

2. Check if the Pods of Tiller and the Kyma Installer are running:

    ```
    kubectl get pods --all-namespaces
    ```

3. To watch the installation progress, run:

    ```
    while true; do \
      kubectl -n default get installation/kyma-installation -o jsonpath="{'Status: '}{.status.state}{', description: '}{.status.description}"; echo; \
      sleep 5; \
    done
    ```

After the installation process is finished, the `Status: Installed, description: Kyma installed` message appears.

If you receive an error, fetch the Kyma Installer logs using this command:

  ```
  kubectl -n kyma-installer logs -l 'name=kyma-installer'
  ```

## Post-installation steps

### Add the xip.io self-signed certificate to your OS trusted certificates

After the installation, add the custom Kyma [`xip.io`](http://xip.io/) self-signed certificate to the trusted certificates of your OS.
For MacOS, run:

  ```
  tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
  && kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.tlsCrt}' | base64 --decode > $tmpfile \
  && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
  && rm $tmpfile
  ```

### Access the cluster

1. To get the address of the cluster's Console, check the host of the Console's virtual service. The name of the host of this virtual service corresponds to the Console URL. To get the virtual service host, run:

    ```
    kubectl get virtualservice core-console -n kyma-system -o jsonpath='{ .spec.hosts[0] }'
    ```

2. Access your cluster under this address:

    ```
    https://{VIRTUAL_SERVICE_HOST}
    ```

3. To log in to your cluster's Console UI, use the default `admin` static user. Click **Login with Email** and sign in with the **admin@kyma.cx** email address. Use the password contained in the `admin-user` Secret located in the `kyma-system` Namespace. To get the password, run:

    ```
    kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode
    ```


  </details>
</div>
