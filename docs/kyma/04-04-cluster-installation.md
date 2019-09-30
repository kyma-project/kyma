---
title: Install Kyma on a cluster
type: Installation
---

This installation guide explains how you can quickly deploy Kyma on a cluster with a wildcard DNS provided by [`xip.io`](http://xip.io) using a GitHub release of your choice.

>**TIP:** A xip.io domain is not recommended for production. If you want to expose the Kyma cluster on your own domain, follow [this](#installation-use-your-own-domain) installation guide. To install Kyma using your own image instead of a GitHub release, follow [these](#installation-use-your-own-kyma-installer-image) instructions.

## Prerequisites

<div tabs name="prerequisites" group="cluster-installation">
  <details>
  <summary label="GKE">
  GKE
  </summary>
  
- [Google Cloud Platform](https://console.cloud.google.com/) (GCP) project with Kubernetes Engine API enabled
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.14.6 or higher
- [gcloud](https://cloud.google.com/sdk/gcloud/)

>**NOTE:** Running Kyma on GKE requires three [`n1-standard-4` machines](https://cloud.google.com/compute/docs/machine-types). You create these machines when you complete the **Prepare the cluster** step.

  </details>
  <details>
  <summary label="AKS">
  AKS
  </summary>

- [Microsoft Azure](https://azure.microsoft.com) account
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.14.6 or higher
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)

>**NOTE:** Running Kyma on AKS requires three [`Standard_D4_v3` machines](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/sizes-general#dv3-series-1). You create these machines when you complete the **Prepare the cluster** step.

  </details>
  <details>
  <summary label="Gardener">
  Gardener
  </summary>

- [Gardener](https://gardener.cloud/) seed cluster
- [Google Cloud Platform](https://console.cloud.google.com/) (GCP) project with Kubernetes Engine API enabled or a [Microsoft Azure](https://azure.microsoft.com) account
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.14.6 or higher

  </details>

</div>

## Choose the release to install

1. Go to [this](https://github.com/kyma-project/kyma/releases/) page and choose the release you want to install.

2. Export the release version as an environment variable. Run:

    ```bash
    export KYMA_VERSION={KYMA_RELEASE_VERSION}
    ```

## Prepare the cluster

<div tabs name="prepare-cluster" group="cluster-installation">
  <details>
  <summary label="GKE">
  GKE
  </summary>
  
1. Select a name for your cluster. Export the cluster name, the name of your GCP project, and the [zone](https://cloud.google.com/compute/docs/regions-zones/) you want to deploy to as environment variables. Run:

    ```bash
    export CLUSTER_NAME={CLUSTER_NAME_YOU_WANT}
    export GCP_PROJECT={YOUR_GCP_PROJECT}
    export GCP_ZONE={GCP_ZONE_TO_DEPLOY_TO}
    ```

2. Create a cluster in the defined zone. Run:

    ```bash
    gcloud container --project "$GCP_PROJECT" clusters \
    create "$CLUSTER_NAME" --zone "$GCP_ZONE" \
    --cluster-version "1.14" --machine-type "n1-standard-4" \
    --addons HorizontalPodAutoscaling,HttpLoadBalancing
    ```

3. Configure kubectl to use your new cluster. Run:

    ```bash
    gcloud container clusters get-credentials $CLUSTER_NAME --zone $GCP_ZONE --project $GCP_PROJECT
    ```

4. Add your account as the cluster administrator:

    ```bash
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)
    ```
  
  </details>
  <details>
  <summary label="AKS">
  AKS
  </summary>

1. Select a name for your cluster. Set the cluster name, the resource group and region as environment variables. Run:

    ```bash
    export RS_GROUP={YOUR_RESOURCE_GROUP_NAME}
    export CLUSTER_NAME={YOUR_CLUSTER_NAME}
    export REGION={YOUR_REGION} #westeurope
    ```

2. Create a resource group for all your resources:

    ```bash
    az group create --name $RS_GROUP --location $REGION
    ```

3. Create an AKS cluster. Run:

    ```bash
    az aks create \
      --resource-group $RS_GROUP \
      --name $CLUSTER_NAME \
      --node-vm-size "Standard_D4_v3" \
      --kubernetes-version 1.14.6 \
      --enable-addons "monitoring,http_application_routing" \
      --generate-ssh-keys
    ```

4. To configure kubectl to use your new cluster, run:

    ```bash
    az aks get-credentials --resource-group $RS_GROUP --name $CLUSTER_NAME
    ```

5. Add additional privileges to be able to access readiness probes endpoints on your AKS cluster.

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/$KYMA_VERSION/installation/resources/azure-crb-for-healthz.yaml
    ```

6. Install custom installation overrides for AKS. Run:

    ```bash
    kubectl create namespace kyma-installer \
    && kubectl create configmap aks-overrides -n kyma-installer --from-literal=global.proxy.excludeIPRanges=10.0.0.1 \
    && kubectl label configmap aks-overrides -n kyma-installer installer=overrides component=istio
    ```

    >**TIP:** An example config map is available [here](./assets/aks-overrides.yaml).

>**CAUTION:** If you define your own Kubernetes jobs on the AKS cluster, follow [this](/components/service-mesh/#troubleshooting-kubernetes-jobs-fail-on-aks) troubleshooting guide to avoid jobs running endlessly on AKS deployments of Kyma.

  </details>
  <details>
  <summary label="Gardener">
  Gardener
  </summary>
  
1. In the left navigation of the Gardener UI, go to the **Secrets** tab and add Secrets to enable provisioning clusters on different architectures. To learn about the requirements for each environment, click the question mark buttons.

2. Provision a cluster form the **Clusters** tab. Click the plus sign in the lower-right corner and choose the infrastructure in which you want to provision your cluster. Apply these settings in the following tabs:

    | Tab  |  Setting |  Required value |
    |---|---|---|
    | Infrastructure |  Kubernetes | `1.14.6`  |
    | Worker  |  Machine type | `n1-standard-4` (GCP) `Standard_D4_v3` (Azure)|
    | Worker  | Autoscaler min.  | `3` |

3. After you provision the cluster, download the kubeconfig file available under the **Show Cluster Access** option in the **Actions** column.

4. Export the downloaded kubeconfig as an environment variable to connect to the cluster you provisioned. Run:

    ```bash
    export KUBECONFIG={PATH_TO_KUBECONFIG_FILE}
    ```

>**NOTE:** If you use an Azure cluster, make sure to run all commands from steps 5 and 6 of [this](#installation-install-kyma-on-a-cluster--prepare-cluster--aks) section.

  </details>
</div>

## Install Kyma

1. Install Tiller on the cluster you provisioned. Run:

   ```bash
   kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/$KYMA_VERSION/installation/resources/tiller.yaml
   ```

2. Deploy Kyma. Run:

    ```bash
    kubectl apply -f https://github.com/kyma-project/kyma/releases/download/$KYMA_VERSION/kyma-installer-cluster.yaml
    ```

3. Check if the Pods of Tiller and the Kyma Installer are running:

    ```bash
    kubectl get pods --all-namespaces
    ```

4. To watch the installation progress, run:

    ```bash
    while true; do \
      kubectl -n default get installation/kyma-installation -o jsonpath="{'Status: '}{.status.state}{', description: '}{.status.description}"; echo; \
      sleep 5; \
    done
    ```

After the installation process is finished, the `Status: Installed, description: Kyma installed` message appears.

If you receive an error, fetch the Kyma Installer logs using this command:

```bash
kubectl -n kyma-installer logs -l 'name=kyma-installer'
```

## Post-installation steps

### Add the xip.io self-signed certificate to your OS trusted certificates

After the installation, add the custom Kyma [`xip.io`](http://xip.io/) self-signed certificate to the trusted certificates of your OS. For MacOS, run:

```bash
  tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
  && kubectl get configmap net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.tlsCrt}' | base64 --decode > $tmpfile \
  && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
  && rm $tmpfile
  ```

### Access the cluster

1. To get the address of the cluster's Console, check the host of the Console's virtual service. The name of the host of this virtual service corresponds to the Console URL. To get the virtual service host, run:

    ```bash
    kubectl get virtualservice core-console -n kyma-system -o jsonpath='{ .spec.hosts[0] }'
    ```

2. Access your cluster under this address:

    ```bash
    https://{VIRTUAL_SERVICE_HOST}
    ```

3. To log in to your cluster's Console UI, use the default `admin` static user. Click **Login with Email** and sign in with the **admin@kyma.cx** email address. Use the password contained in the `admin-user` Secret located in the `kyma-system` Namespace. To get the password, run:

    ```bash
    kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode
    ```

If you need to use Helm to manage your Kubernetes resources, complete the [additional configuration](#installation-use-helm) after you finish the installation.
