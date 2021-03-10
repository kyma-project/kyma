---
title: Install Kyma on a cluster
type: Installation
---

This installation guide explains how you can quickly deploy Kyma on a cluster with a wildcard DNS provided by [xip.io](http://xip.io) using a GitHub release of your choice.

>**TIP:** The xip.io domain is not recommended for production. If you want to expose the Kyma cluster on your own domain, follow the [installation guide](#installation-install-kyma-with-your-own-domain). To install Kyma using your own image instead of a GitHub release, follow the [instructions](#installation-use-your-own-kyma-installer-image).

## Prerequisites

>**CAUTION:** Kubernetes is [deprecating Docker](https://kubernetes.io/blog/2020/12/02/dont-panic-kubernetes-and-docker/) as a container runtime after v1.20 and [we are working on compatibility](https://github.com/kyma-project/kyma/issues/10842) all of our components with the `containerD`. Be sure that the kyma is not installed on a cluster with the `containerD` and the `self-sign certificate` at the same time. If yes, upgrade the cluster to use the `docker` instead of the `containerD` or [generate the tls certificate](https://kyma-project.io/docs/#installation-install-kyma-with-your-own-domain-generate-the-tls-certificate).

<div tabs name="prerequisites" group="cluster-installation">
  <details>
  <summary label="GKE">
  GKE
  </summary>

- Kubernetes cluster v1.18
- [Kyma CLI](https://github.com/kyma-project/cli)
- [Google Cloud Platform](https://console.cloud.google.com/) (GCP) project with Kubernetes Engine API enabled
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.16.3 or higher
- [gcloud](https://cloud.google.com/sdk/gcloud/)

>**NOTE:** Running Kyma on GKE requires three [`n1-standard-4` machines](https://cloud.google.com/compute/docs/machine-types). The Kyma production profile requires at least `n1-standard-8` machines, but it is recommended to use the `c2-standard-8` type. Create these machines when you complete the **Prepare the cluster** step.

  </details>
  <details>
  <summary label="AKS">
  AKS
  </summary>

- Kubernetes cluster v1.19
- [Kyma CLI](https://github.com/kyma-project/cli)
- [Microsoft Azure](https://azure.microsoft.com) account
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.16.3 or higher
- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)

>**NOTE:** Running Kyma on AKS requires three [`Standard_D4_v3` machines](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes-general). The Kyma production profile requires at least `Standard_F8s_v2` machines, but it is recommended to use the `Standard_D8_v3` type. Create these machines when you complete the **Prepare the cluster** step.

  </details>
  <details>
  <summary label="Gardener">
  Gardener
  </summary>

- Kubernetes cluster v1.19
- [Kyma CLI](https://github.com/kyma-project/cli)
- [Gardener](https://gardener.cloud/) account
- [Google Cloud Platform](https://console.cloud.google.com/) (GCP) project
- [Microsoft Azure](https://azure.microsoft.com) project
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.16.3 or higher

  </details>
</div>

## Choose the release to install

1. Go to [Kyma releases](https://github.com/kyma-project/kyma/releases/) and choose the release you want to install.

2. Export the release version as an environment variable:

    ```bash
    export KYMA_VERSION={KYMA_RELEASE_VERSION}
    ```

## Prepare the cluster

<div tabs name="prepare-cluster" group="cluster-installation">
  <details>
  <summary label="GKE">
  GKE
  </summary>

1. Create a service account and a service account key as JSON following [these steps](https://github.com/kyma-incubator/hydroform/blob/master/provision/examples/gcp/README.md#configure-gcp).

2. Export the cluster name, the name of your GCP project, and the [zone](https://cloud.google.com/compute/docs/regions-zones/) you want to deploy to as environment variables:

    ```bash
    export CLUSTER_NAME={CLUSTER_NAME_YOU_WANT}
    export GCP_PROJECT={YOUR_GCP_PROJECT}
    export GCP_ZONE={GCP_ZONE_TO_DEPLOY_TO}
    ```

3. Create a cluster in the defined zone:

    ```bash
    kyma provision gke -c {SERVICE_ACCOUNT_KEY_FILE_PATH} -n $CLUSTER_NAME -l $GCP_ZONE -p $GCP_PROJECT
    ```
   >**NOTE**: Kyma offers the production profile. Pass the `-t` flag to Kyma CLI with the `n1-standard-8` or `c2-standard-8` value if you want to use it.

4. Configure kubectl to use your new cluster:

    ```bash
    gcloud container clusters get-credentials $CLUSTER_NAME --zone $GCP_ZONE --project $GCP_PROJECT
    ```

5. Add your account as the cluster administrator:

    ```bash
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)
    ```

  </details>
  <details>
  <summary label="AKS">
  AKS
  </summary>

1. Select a name for your cluster. Set the cluster name, the resource group and region as environment variables:

    ```bash
    export RS_GROUP={YOUR_RESOURCE_GROUP_NAME}
    export CLUSTER_NAME={YOUR_CLUSTER_NAME}
    export REGION={YOUR_REGION} #westeurope
    ```

2. Create a resource group for all your resources:

    ```bash
    az group create --name $RS_GROUP --location $REGION
    ```

3. Create a [service principal](https://docs.microsoft.com/en-us/azure/aks/kubernetes-service-principal#manually-create-a-service-principal) on Azure. Create a JSON file with the Azure Client ID, Client Secret, Subscription ID, and Tenant ID:

    ```json
    {
      "subscription_id": "{YOUR_SUBSCRIPTION_ID}",
      "tenant_id": "{YOUR_TENANT_ID}",
      "client_id": "{YOUR_APP_ID}",
      "client_secret": "{YOUR_APP_PASSWORD}"
    }
    ```

4. Create an AKS cluster:

    ```bash
    kyma provision aks -c {YOUR_CREDENTIALS_FILE_PATH} -n $CLUSTER_NAME -p $RS_GROUP -l $REGION
    ```
   >**NOTE**: Kyma offers the production profile. Pass the flag `-t` to Kyma CLI with `Standard_F8s_v2` or `Standard_D8_v3` if you want to use it.

5. Add additional privileges to be able to access readiness probes endpoints on your AKS cluster:

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/$KYMA_VERSION/installation/resources/azure-crb-for-healthz.yaml
    ```
  >**CAUTION:** If you define your own Kubernetes jobs on the AKS cluster, follow the [troubleshooting guide](/components/service-mesh/#troubleshooting-kubernetes-jobs-fail-on-aks) to avoid jobs running endlessly on AKS deployments of Kyma.

  </details>
  <details>
  <summary label="Gardener">
  Gardener
  </summary>

1. Use the Gardener dashboard to configure provider settings.

    >**NOTE:** You need to perform these steps only once.

    * For GCP:
      * Create a project in Gardener.
      * Add a [new service account and roles](https://gardener.cloud/documentation/guides/administer_shoots/gardener_gcp).
      * Add the GCP Secret under **Secrets** in the Gardener dashboard.
      * Add the service account and download Gardener's `kubeconfig` file.

    * For Azure:
      * Create a project in Gardener.
      * Add the Azure Secret under **Secrets** in the Gardener dashboard. Use the details of your Azure service account. If do not have an account, request one.
      * Add the service account and download Gardener's `kubeconfig` file.

2. Provision the cluster using the [Kyma CLI](https://github.com/kyma-project/cli).

   >**NOTE**: Kyma offers the [production profile](/components/service-mesh/#configuration-service-mesh-production-profile) which requires a different machine type. Specify it using the `--type` flag.

   To provision a Gardener cluster on GCP, run:

   ```
   kyma provision gardener gcp -n {cluster_name} -p {project_name} -s {kyma_gardener_gcp_secret_name} -c {path_to_gardener_kubeconfig}
   ```
   See the complete [list of flags and their descriptions](https://github.com/kyma-project/cli/blob/master/docs/gen-docs/kyma_provision_gardener_gcp.md).

   To provision a Gardener cluster on Azure, run:

   ```
   kyma provision gardener az -n {cluster_name} -p {project_name} -s {kyma_gardener_azure_secret_name} -c {path_to_gardener_kubeconfig}
   ```
   See the complete [list of flags and their descriptions](https://github.com/kyma-project/cli/blob/master/docs/gen-docs/kyma_provision_gardener_az.md).

3. After you provision the cluster, its `kubeconfig` file will be downloaded and automatically set as the current context.

  </details>
</div>

## Install Kyma

Install Kyma using Kyma CLI:

```bash
kyma install -s $KYMA_VERSION
```

To install Kyma with one of the predefined [profiles](#installation-overview-profiles), run:

```bash
kyma install -s $KYMA_VERSION --profile {evaluation|production}
```

>**NOTE:** If you don't specify `$KYMA_VERSION`, the version from the latest commit on the `master` branch is installed. If you don't specify the profile, the default version of Kyma is installed.

## Post-installation steps

### Access the cluster

1. To open the cluster's Console on your default browser, run:

    ```bash
    kyma console
    ```

2. To log in to your cluster's Console UI, use the default `admin` static user. Click **Login with Email** and sign in with the **admin@kyma.cx** email address. Use the password printed after the installation. To get the password manually, you can also run:

    ```bash
    kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode
    ```

If you need to use Helm to manage your Kubernetes resources, read the [additional configuration](#installation-use-helm) document.
