---
title: Use your own domain
type: Installation
---

This guide explains how to deploy Kyma on a cluster using your own domain.

>**TIP:** Get a free domain for your cluster using services like [freenom.com](https://www.freenom.com) or similar.

Choose your cloud provider and get started:

<div tabs name="provider-domain">
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
- A domain for your GKE cluster

>**NOTE:** Running Kyma on GKE requires three [`n1-standard-4` machines](https://cloud.google.com/compute/docs/machine-types). You create these machines when you complete the **Prepare the GKE cluster** step.

## Choose the release to install

1. Go to [this](https://github.com/kyma-project/kyma/releases/) page and choose the release you want to install.

2. Export the release version as an environment variable. Run:
    ```
    export KYMA_VERSION={KYMA_RELEASE_VERSION}
    ```

## DNS setup and TLS certificate generation

### Delegate the management of your domain to Google Cloud DNS

> **NOTE**: Google Cloud DNS setup has to be done only once per DNS zone

Follow these steps:

1. Export the project name, domain name, and DNS zone name as environment variables. Run the commands listed below:

    ```
    export GCP_PROJECT={YOUR_GCP_PROJECT}
    export DNS_NAME={YOUR_ZONE_DOMAIN}
    export DNS_ZONE={YOUR_DNS_ZONE}
    ```
    Example:
    ```
    export DNS_NAME=kyma-demo.ga
    export DNS_ZONE=myzone
    ```

2. Create a DNS-managed zone in your Google project. Run:

    ```
    gcloud dns --project=$GCP_PROJECT managed-zones create $DNS_ZONE --description= --dns-name=$DNS_NAME
    ```

    Alternatively, create it through the GCP UI. Navigate go to **Network Services** in the **Network** section, click **Cloud DNS** and select **Create Zone**.

3. Delegate your domain to Google name servers.

    - Get the list of the name servers from the zone details. This is a sample list:
      ```
      ns-cloud-b1.googledomains.com.
      ns-cloud-b2.googledomains.com.
      ns-cloud-b3.googledomains.com.
      ns-cloud-b4.googledomains.com.
      ```

    - Set up your domain to use these name servers.

4. Check if everything is set up correctly and your domain is managed by Google name servers. Run:
    ```
    host -t ns $DNS_NAME
    ```
    A successful response returns the list of the name servers you fetched from GCP.

### Get the TLS certificate

1. Export the certificate issuer email and the cluster domain as environment variables:
    ```
    export CERT_ISSUER_EMAIL={YOUR_EMAIL}
    export DOMAIN="$CLUSTER_NAME.$(echo $DNS_NAME | sed `s/\.$//`)"
    ```

2. Create a folder for certificates. Run:
    ```
    mkdir letsencrypt
    ```

3. Create a new service account and assign it to the **dns.admin** role. Run these commands:
    ```
    gcloud iam service-accounts create dnsmanager --display-name "dnsmanager" --project "$GCP_PROJECT"
    ```
    ```
    gcloud projects add-iam-policy-binding $GCP_PROJECT \
        --member serviceAccount:dnsmanager@$GCP_PROJECT.iam.gserviceaccount.com --role roles/dns.admin
    ```

    > **NOTE**: You don't have to create a new DNS manager service account (SA) every time you deploy a cluster. Instead, you can use an existing SA that has the **dns.admin** assigned.


4. Generate an access key for this account in the `letsencrypt` folder. Run:
    ```
    gcloud iam service-accounts keys create ./letsencrypt/key.json --iam-account dnsmanager@$GCP_PROJECT.iam.gserviceaccount.com
    ```

    > **NOTE**: The number of keys you can generate for a single service account is limited. Reuse the existing keys instead of generating a new key for every cluster.

5. Run the Certbot Docker image with the `letsencrypt` folder mounted. Certbot uses the key to apply DNS challenge for the certificate request and stores the TLS certificates in that folder. Run:
    ```
    docker run -it --name certbot --rm \
        -v "$(pwd)/letsencrypt:/etc/letsencrypt" \
        certbot/dns-google \
        certonly \
        -m $CERT_ISSUER_EMAIL --agree-tos --no-eff-email \
        --dns-google \
        --dns-google-credentials /etc/letsencrypt/key.json \
        --server https://acme-v02.api.letsencrypt.org/directory \
        -d "*.$DOMAIN"
    ```

6. Export the certificate and key as environment variables. Run these commands:

    ```
    export TLS_CERT=$(cat ./letsencrypt/live/$DOMAIN/fullchain.pem | base64 | sed 's/ /\\ /g' | tr -d '\n');
    export TLS_KEY=$(cat ./letsencrypt/live/$DOMAIN/privkey.pem | base64 | sed 's/ /\\ /g' | tr -d '\n')
    ```

## Prepare the GKE cluster

1. Select a name for your cluster. Set the cluster name and the zone you want to deploy to as environment variables. Run:
    ```
    export CLUSTER_NAME={CLUSTER_NAME_YOU_WANT}
    export GCP_ZONE={GCP_ZONE_TO_DEPLOY_TO}
    ```

2. Create a cluster in the configured zone. Run:
    ```
    gcloud container --project "$GCP_PROJECT" clusters \
    create "$CLUSTER_NAME" --zone "$GCP_ZONE" \
    --cluster-version "1.12" --machine-type "n1-standard-4" \
    --addons HorizontalPodAutoscaling,HttpLoadBalancing
    ```

3. Ensure kubectl is configured to use your new cluster. Run:
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

6. Install custom installation overrides for your DNS domain and TLC certifcates. Run:
    ```
    kubectl create namespace kyma-installer \
    && kubectl create configmap owndomain-overrides -n kyma-installer --from-literal=global.domainName=$DOMAIN --from-literal=global.tlsCrt=$TLS_CERT --from-literal=global.tlsKey=$TLS_KEY \
    && kubectl label configmap owndomain-overrides -n kyma-installer installer=overrides
    ```

    >**TIP:** An example config map is available [here](./assets/owndomain-overrides.yaml)

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
In case of an error, you can fetch the logs from the Kyma Installer by running:
    ```
    kubectl -n kyma-installer logs -l 'name=kyma-installer'
    ```

## Configure DNS for the cluster load balancer

To add DNS entries, run these commands:
```
export EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

export APISERVER_PUBLIC_IP=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

gcloud dns --project=$PROJECT record-sets transaction start --zone=$DNS_ZONE

gcloud dns --project=$PROJECT record-sets transaction add $EXTERNAL_PUBLIC_IP --name=\*.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE

gcloud dns --project=$PROJECT record-sets transaction add $APISERVER_PUBLIC_IP --name=\apiserver.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE

gcloud dns --project=$PROJECT record-sets transaction execute --zone=$DNS_ZONE
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
- A domain for your AKS cluster


>**NOTE:** Running Kyma on AKS requires three [`Standard_D4_v3` machines](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/sizes-general#dv3-series-1). You create these machines when you complete the **Prepare the AKS cluster** step.

>**CAUTION:** Due to a known Istio-related issue, Kubernetes jobs run endlessly on AKS deployments of Kyma. Read [this](/components/service-mesh/#troubleshooting-kubernetes-jobs-fail-on-aks) document to learn more.

## Choose the release to install

1. Go to [this](https://github.com/kyma-project/kyma/releases/) page and choose the release you want to install.

2. Export the release version as an environment variable. Run:
    ```
    export KYMA_VERSION={KYMA_RELEASE_VERSION}
    ```

## DNS setup and TLS certificate generation

### Delegate the management of your domain to Azure DNS

Follow these steps:

1. Export the domain name, the sub-domain, and the resource group name as environment variables. Run these commands:

    ```
    export DNS_DOMAIN={YOUR_DOMAIN} # example.com
    export SUB_DOMAIN={YOUR_SUBDOMAIN} # cluster (in this case the full name of your cluster is cluster.example.com)
    export DOMAIN="$SUB_DOMAIN.$DNS_DOMAIN" # cluster.example.com
    export RS_GROUP={YOUR_RESOURCE_GROUP_NAME}
    ```

2. Create a DNS-managed zone in your Azure subscription. Run:

    ```
    az network dns zone create -g $RS_GROUP -n $DNS_DOMAIN
    ```

    Alternatively, create it through the Azure UI. In the **Networking** section, go to **All services**, click **DNS zones**, and select **Add**.

3. Delegate your domain to Azure name servers.

    - Get the list of the name servers from the zone details. This is a sample list:
      ```
      ns1-05.azure-dns.com.
      ns2-05.azure-dns.net.
      ns3-05.azure-dns.org.
      ns4-05.azure-dns.info.
      ```

    - Set up your domain to use these name servers.

4. Check if everything is set up correctly and your domain is managed by Azure name servers. Run:
    ```
    host -t ns $DNS_DOMAIN
    ```
    A successful response returns the list of the name servers you fetched from Azure.

### Get the TLS certificate

>**NOTE:** Azure DNS is not yet supported by Certbot so you must perform manual verification.

1. Create a folder for certificates. Run:
    ```
    mkdir letsencrypt
    ```
2. Export your email address as an environment variable:
    ```
    export YOUR_EMAIL={YOUR_EMAIL}
    ```
3. To get the certificate, run the Certbot Docker image with the `letsencrypt` folder mounted. Certbot stores the TLS certificates in that folder.
    ```
    docker run -it --name certbot --rm \
        -v "$(pwd)/letsencrypt:/etc/letsencrypt" \
        certbot/certbot \
        certonly \
        -m $YOUR_EMAIL --agree-tos --no-eff-email \
        --manual \
        --manual-public-ip-logging-ok \
        --preferred-challenges dns \
        --server https://acme-v02.api.letsencrypt.org/directory \
        -d "*.$SUB_DOMAIN.$DNS_DOMAIN"
    ```
    You will see the following message:

    ```
    Please deploy a DNS TXT record under the name
    _acme-challenge.rc2-test.kyma.online with the following value:

    # TXT_VALUE

    Before continuing, verify the record is deployed.
    ```
    Copy the `TXT_VALUE`.

3. Open a new terminal and export these environment variables:
    ```
    export DNS_DOMAIN={YOUR_DOMAIN} # example.com
    export SUB_DOMAIN={YOUR_SUBDOMAIN} # cluster (in this case the full name of your cluster is cluster.example.com)
    export RS_GROUP={YOUR_RESOURCE_GROUP_NAME}
    ```

4. Export the `TXT_VALUE`.

    ```
    export TXT_VALUE={YOUR_TXT_VALUE}
    ```
    To modify TXT record for your domain, run:
    ```
    az network dns record-set txt delete -n "_acme-challenge.$SUB_DOMAIN" -g $RS_GROUP -z $DNS_DOMAIN --yes
    az network dns record-set txt create -n "_acme-challenge.$SUB_DOMAIN" -g $RS_GROUP -z $DNS_DOMAIN --ttl 60 > /dev/null
    az network dns record-set txt add-record -n "_acme-challenge.$SUB_DOMAIN" -g $RS_GROUP -z $DNS_DOMAIN --value $TXT_VALUE
    ```
5. Go back to the first console, wait about 2 minutes and press enter.

6. Export the certificate and key as environment variables. Run these commands:

    ```
    export TLS_CERT=$(cat ./letsencrypt/live/$SUB_DOMAIN.$DNS_DOMAIN/fullchain.pem | base64 | sed 's/ /\\ /g')
    export TLS_KEY=$(cat ./letsencrypt/live/$SUB_DOMAIN.$DNS_DOMAIN/privkey.pem | base64 | sed 's/ /\\ /g')
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

6. Install custom installation overrides for AKS, your DNS domain and TLC certifcates. Run:
    ```
    kubectl create namespace kyma-installer \
    && kubectl create configmap owndomain-overrides -n kyma-installer --from-literal=global.domainName=$DOMAIN --from-literal=global.tlsCrt=$TLS_CERT --from-literal=global.tlsKey=$TLS_KEY \
    && kubectl label configmap owndomain-overrides -n kyma-installer installer=overrides \
    && kubectl create configmap aks-overrides -n kyma-installer --from-literal=global.proxy.excludeIPRanges=10.0.0.1 \
    && kubectl label configmap aks-overrides -n kyma-installer installer=overrides component=istio
    ```

    >**TIP:** An example config map is available [here](./assets/owndomain-overrides.yaml) and [here](./assets/aks-overrides.yaml)

## Install Kyma

1. Deploy Kyma. Run:
    ```
    kubectl apply -f https://github.com/kyma-project/kyma/releases/download/$LATEST/kyma-installer-cluster.yaml
    ```
    >**NOTE:** If you get `Error from server (MethodNotAllowed)`, run the command again before proceeding to the next step.

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
    In case of an error, you can fetch the logs from the Kyma Installer by running:
    ```
    kubectl -n kyma-installer logs -l 'name=kyma-installer'
    ```

## Configure DNS for the cluster load balancer

Run these commands:

```
export EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

export APISERVER_PUBLIC_IP=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

az network dns record-set a create -g $RS_GROUP -z $DNS_DOMAIN -n \*.$SUB_DOMAIN --ttl 60
az network dns record-set a add-record -g $RS_GROUP -z $DNS_DOMAIN -n \*.$SUB_DOMAIN -a $EXTERNAL_PUBLIC_IP

az network dns record-set a create -g $RS_GROUP -z $DNS_DOMAIN -n apiserver.$SUB_DOMAIN --ttl 60
az network dns record-set a add-record -g $RS_GROUP -z $DNS_DOMAIN -n apiserver.$SUB_DOMAIN -a $APISERVER_PUBLIC_IP
```


</details>
</div>

## Access the cluster

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
