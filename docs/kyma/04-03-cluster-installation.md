---
title: Install Kyma on a cluster
type: Installation
---

This installation guide explains how you can quickly deploy Kyma on a cluster with a wildcard DNS provided by [`xip.io`](http://xip.io) using a GitHub release of your choice.

>**NOTE:** If you want to expose the Kyma cluster on your own domain, follow [this](#installation-use-your-own-domain) installation guide. To install using your own image instead of a GitHub release, follow [these](#installation-use-your-own-kyma-installer-image) instructions.

If you need to use Helm and access Tiller, complete the [additional configuration](#installation-use-helm) after the installation.

>**CAUTION:** These instructions are valid starting with Kyma 1.2. If you want to install older releases, refer to the respective documentation versions.

Choose your cloud provider and get started:

<div tabs>
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
      --node-vm-size "Standard_DS2_v2" \
      --kubernetes-version 1.10.9 \
      --enable-addons "monitoring,http_application_routing" \
      --generate-ssh-keys
    ```

4. To configure kubectl to use your new cluster, run:
    ```
    az aks get-credentials --resource-group $RS_GROUP --name $CLUSTER_NAME
    ```

5. Install Tiller and add additional privileges to be able to access readiness probes endpoints on your AKS cluster.

    * Installation from release
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

    >**TIP:** An example config map is available [here](./assets/aks-overrides.yaml)
  </details>
</div>

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
In case of an error, you can fetch the logs from the Installer by running:
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
