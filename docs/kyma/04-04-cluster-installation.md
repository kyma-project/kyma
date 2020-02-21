---
title: Install Kyma on a cluster
type: Installation
---

This installation guide explains how you can quickly deploy Kyma on a cluster with a wildcard DNS provided by [`xip.io`](http://xip.io) using a GitHub release of your choice.

>**TIP:** An xip.io domain is not recommended for production. If you want to expose the Kyma cluster on your own domain, follow [this](#installation-use-your-own-domain) installation guide. To install Kyma using your own image instead of a GitHub release, follow [these](#installation-use-your-own-kyma-installer-image) instructions.

## Prerequisites

<div tabs name="prerequisites" group="cluster-installation">
  <details>
  <summary label="GKE">
  GKE
  </summary>

- [Google Cloud Platform](https://console.cloud.google.com/) (GCP) project with Kubernetes Engine API enabled
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.16.3 or higher
- [gcloud](https://cloud.google.com/sdk/gcloud/)

>**NOTE:** Running Kyma on GKE requires three [`n1-standard-4` machines](https://cloud.google.com/compute/docs/machine-types). You create these machines when you complete the **Prepare the cluster** step.

  </details>
  <details>
  <summary label="AKS">
  AKS
  </summary>

- [Microsoft Azure](https://azure.microsoft.com) account
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.16.3 or higher
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)

>**NOTE:** Running Kyma on AKS requires three [`Standard_D4_v3` machines](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes-general). You create these machines when you complete the **Prepare the cluster** step.

  </details>
  <details>
  <summary label="Gardener">
  Gardener
  </summary>

- [Gardener](https://gardener.cloud/) account
- [Google Cloud Platform](https://console.cloud.google.com/) (GCP) project
- [Microsoft Azure](https://azure.microsoft.com) project
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.16.3 or higher

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
    --cluster-version "1.15" --machine-type "n1-standard-4" \
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
      --kubernetes-version 1.15.7 \
      --enable-addons "monitoring,http_application_routing" \
      --generate-ssh-keys \
      --max-pods 110
    ```

4. To configure kubectl to use your new cluster, run:

    ```bash
    az aks get-credentials --resource-group $RS_GROUP --name $CLUSTER_NAME
    ```

5. Add additional privileges to be able to access readiness probes endpoints on your AKS cluster.

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/$KYMA_VERSION/installation/resources/azure-crb-for-healthz.yaml
    ```

>**CAUTION:** If you define your own Kubernetes jobs on the AKS cluster, follow [this](/components/service-mesh/#troubleshooting-kubernetes-jobs-fail-on-aks) troubleshooting guide to avoid jobs running endlessly on AKS deployments of Kyma.

  </details>
  <details>
  <summary label="Gardener">
  Gardener
  </summary>

1. Use the Gardener dashboard to configure provider settings.

    >**NOTE:** You need to perform these steps only once.
   
    * For GCP:
      * Create a project in Gardener.
      * Add a [new service account and roles](https://gardener.cloud/050-tutorials/content/howto/gardener_gcp/#create-a-new-serviceaccount-and-assign-roles).
      * Add the GCP Secret under **Secrets** in the Gardener dashboard.
      * Add the service account and download Gardener's `kubeconfig` file.

    * For AKS:
      * Create a project in Gardener.
      * Add the Azure Secret under **Secrets** in the Gardener dashboard. Use the details of your Azure service account. If do not have an account, request one.
      * Add the service account and download Gardener's `kubeconfig` file.

2. Provision the cluster using the [Kyma CLI](https://github.com/kyma-project/cli).

   To provision a GKE cluster, run:

   ```
   kyma provision gardener -n {cluster_name} -p {project_name} -s {kyma_gardener_gcp_secret_name} -c {path_to_gardener_kubeconfig}
   ```

   To provision an AKS cluster, run:

   ```
   kyma provision gardener --target-provider azure -n {cluster_name} -p {project_name} -s {kyma_gardener_azure_secret_name} -c {path_to_gardener_kubeconfig} -t Standard_D2_v3 --region westeurope --disk-size 35 --disk-type Standard_LRS --extra vnetcidr="10.250.0.0/19"
   ```
   For a complete list of flags and their descriptions, see [this](https://github.com/kyma-project/cli/blob/master/docs/gen-docs/kyma_provision_gardener.md) document.

3. After you provision the cluster, its `kubeconfig` file will be downloaded and automatically set as the current context. 


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
