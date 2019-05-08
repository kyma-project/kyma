---
title: Install Kyma on a cluster
type: Installation
---

This installation guide explains how you can quickly deploy Kyma on a cluster with a wildcard DNS provided by [`xip.io`](http://xip.io).

>**NOTE:** If you have your own domain and want to use it during installation, follow [this](#installation-use-your-own-domain) guide.

Kyma cluster installation comes down to these steps:

1. Preparation of the cluster
2. Preparation of the configuration file
3. Kyma deployment

The guide explains how to prepare the configuration file from the latest GitHub release. If you want to use your own image, follow [these](#installation-use-your-own-kyma-installer-image) steps.
Additionally, if you need to use Helm and access Tiller securely, complete [additional configuration](#installation-use-helm) at the end of the installation procedure.

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

## Installation process

### Prepare the GKE cluster

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
    --cluster-version "1.12.5" --machine-type "n1-standard-4" \
    --addons HorizontalPodAutoscaling,HttpLoadBalancing
    ```

3. Add your account as the cluster administrator:
    ```
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)
    ```

### Prepare the configuration file

Use the GitHub release 0.8 or higher.

1. Go to [this](https://github.com/kyma-project/kyma/releases/) page and choose the release you want to install.

2. Export the release version as an environment variable. Run:
    ```
    export KYMA_VERSION={KYMA_RELEASE_VERSION}
    ```
 
3. Install Tiller on your GKE cluster. Run:

    ```
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/$KYMA_VERSION/installation/resources/tiller.yaml
    ```

4. Download the `kyma-config-cluster.yaml` and `kyma-installer-cluster.yaml` files from the latest release. Run:
   ```
   wget https://github.com/kyma-project/kyma/releases/download/$KYMA_VERSION/kyma-config-cluster.yaml
   wget https://github.com/kyma-project/kyma/releases/download/$KYMA_VERSION/kyma-installer-cluster.yaml
   ```

5. Prepare the deployment file.

    - Run this command:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Alternatively, run this command if you deploy Kyma with GKE version 1.12.6-gke.X and above:

    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

6. The output of this operation is the `my-kyma.yaml` file. Use it to deploy Kyma on your GKE cluster.


### Deploy Kyma

1. Configure kubectl to use your new cluster. Run:
    ```
    gcloud container clusters get-credentials $CLUSTER_NAME --zone $GCP_ZONE --project $GCP_PROJECT
    ```

2. Deploy Kyma using the `my-kyma` custom configuration file you created. Run:
    ```
    kubectl apply -f my-kyma.yaml
    ```

3. Check if the Pods of Tiller and the Kyma Installer are running:
    ```
    kubectl get pods --all-namespaces
    ```

4. Start Kyma installation:
    ```
    kubectl label installation/kyma-installation action=install
    ```

5. To watch the installation progress, run:
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


  </details>
  <details>
  <summary>
  AKS
  </summary>


Install Kyma on an [Azure Kubernetes Service](https://azure.microsoft.com/services/kubernetes-service/) (AKS) cluster.

## Prerequisites

- [Microsoft Azure](https://azure.microsoft.com)
- [Kubernetes](https://kubernetes.io/) 1.12
- Tiller 2.10.0 or higher
- [Docker](https://www.docker.com/)
- [Docker Hub](https://hub.docker.com/) account
- [az](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)

## Installation process

### Prepare the AKS cluster

Set the following environment variables:
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

### Prepare the configuration file

Use the GitHub release 0.8 or higher.

1. Go to [this](https://github.com/kyma-project/kyma/releases/) page and choose the latest release.

2. Export the release version as an environment variable. Run:
    ```
    export LATEST={KYMA_RELEASE_VERSION}
    ```

3. Download the `kyma-config-cluster.yaml` and `kyma-installer-cluster.yaml` files from the latest release. Run:
   ```
   wget https://github.com/kyma-project/kyma/releases/download/$LATEST/kyma-config-cluster.yaml
   wget https://github.com/kyma-project/kyma/releases/download/$LATEST/kyma-installer-cluster.yaml
   ```

4. Prepare the deployment file.

    - Run this command:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Alternatively, run this command if you deploy Kyma with Kubernetes version 1.14 and above:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

5. The output of this operation is the `my_kyma.yaml` file. Use it to deploy Kyma on your GKE cluster.

### Deploy Kyma

1. Deploy Kyma using the `my-kyma` custom configuration file you created. Run:
    ```
    kubectl apply -f my-kyma.yaml
    ```
    >**NOTE:** If you get `Error from server (MethodNotAllowed)`, run the command again before proceeding to the next step.

2. Check if the Pods of Tiller and the Kyma Installer are running:
    ```
    kubectl get pods --all-namespaces
    ```

3. Start Kyma installation:
    ```
    kubectl label installation/kyma-installation action=install
    ```

4. To watch the installation progress, run:
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


  </details>
</div>

## Post-installation steps

### Add the xip.io self-signed certificate to your OS trusted certificates

After the installation, add the custom Kyma [`xip.io`](http://xip.io/) self-signed certificate to the trusted certificates of your OS. For MacOS, run:
```
tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
&& kubectl get configmap cluster-certificate-overrides -n kyma-installer -o jsonpath='{.data.global\.tlsCrt}' | base64 --decode > $tmpfile \
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
