---
title: Install Kyma with your own domain
type: Installation
---

This guide explains how to deploy Kyma on a cluster using your own domain.

>**TIP:** Get a free domain for your cluster using services like [freenom.com](https://www.freenom.com) or similar.

## Prerequisites

<div tabs name="prerequisites" group="use-your-own-domain">
  <details>
  <summary label="GKE">
  GKE
  </summary>

- A domain for your [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/) (GKE) cluster
- [Kyma CLI](https://github.com/kyma-project/cli)
- [Google Cloud Platform](https://console.cloud.google.com/) (GCP) project with Kubernetes Engine API enabled
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.16.3 or higher
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- [wget](https://www.gnu.org/software/wget/)

>**NOTE:** Running Kyma on GKE requires three [`n1-standard-4` machines](https://cloud.google.com/compute/docs/machine-types). The Kyma production profile requires at least `n1-standard-8` machines, but it is recommended to use the `c2-standard-8` type. Create these machines when you complete the **Prepare the cluster** step.

  </details>
  <details>
  <summary label="AKS">
  AKS
  </summary>

- A domain for your [Azure Kubernetes Service](https://azure.microsoft.com/services/kubernetes-service/) (AKS) cluster
- [Kyma CLI](https://github.com/kyma-project/cli)
- [Microsoft Azure](https://azure.microsoft.com)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.16.3 or higher
- [Docker](https://www.docker.com/)
- A [Docker Hub](https://hub.docker.com/) account
- [az](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)

>**NOTE:** Running Kyma on AKS requires three [`Standard_D4_v3` machines](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes-general). The Kyma production profile requires at least `Standard_F8s_v2` machines, but it is recommended to use the `Standard_D8_v3` type. Create these machines when you complete the **Prepare the cluster** step.

  </details>

</div>

## Choose the release to install

1. Go to [Kyma releases](https://github.com/kyma-project/kyma/releases/) and choose the release you want to install.

2. Export the release version as an environment variable. Run:

    ```bash
    export KYMA_VERSION={KYMA_RELEASE_VERSION}
    ```

## Set up the DNS

<div tabs name="dns-setup" group="use-your-own-domain">
  <details>
  <summary label="GKE">
  GKE
  </summary>

Delegate the management of your domain to Google Cloud DNS.

>**NOTE**: Google Cloud DNS needs to be set up only once per a DNS zone.

Follow these steps:

1. Export the project name, the domain name, and the DNS zone name as environment variables. Run these commands:

    ```bash
    export GCP_PROJECT={YOUR_GCP_PROJECT}
    export DNS_NAME={YOUR_ZONE_DOMAIN}
    export DNS_ZONE={YOUR_DNS_ZONE}
    ```

2. Create a DNS-managed zone in your Google project. Run:

    ```bash
    gcloud dns --project=$GCP_PROJECT managed-zones create $DNS_ZONE --description= --dns-name=$DNS_NAME
    ```

    Alternatively, create the DNS-managed zone through the GCP UI. In the **Network** section navigate to **Network Services**, click **Cloud DNS**, and select **Create Zone**.

3. Delegate your domain to Google name servers.

    - Get the list of the name servers from the zone details. This is a sample list:

    ```bash
    ns-cloud-b1.googledomains.com.
    ns-cloud-b2.googledomains.com.
    ns-cloud-b3.googledomains.com.
    ns-cloud-b4.googledomains.com.
    ```

    - Set up your domain to use these name servers.

4. Check if everything is set up correctly and your domain is managed by Google name servers. Run:

   ```bash
   host -t ns $DNS_NAME
   ```

   A successful response returns the list of the name servers you fetched from GCP.

</details>
  <details>
  <summary label="AKS">
  AKS
  </summary>

Delegate the management of your domain to Azure DNS. Follow these steps:

1. Export the domain name, the sub-domain, and the resource group name as environment variables. Run these commands:

    ```bash
    export DNS_DOMAIN={YOUR_DOMAIN} # example.com
    export SUB_DOMAIN={YOUR_SUBDOMAIN} # cluster (in this case the full name of your cluster is cluster.example.com)
    export DOMAIN="$SUB_DOMAIN.$DNS_DOMAIN" # cluster.example.com
    export RS_GROUP={YOUR_RESOURCE_GROUP_NAME}
    ```

2. Create a DNS-managed zone in your Azure subscription. Run:

    ```bash
    az network dns zone create -g $RS_GROUP -n $DNS_DOMAIN
    ```

    Alternatively, create it through the Azure UI. In the **Networking** section, go to **All services**, click **DNS zones**, and select **Add**.

3. Delegate your domain to Azure name servers.

    - Get the list of the name servers from the zone details. This is a sample list:

    ```bash
    ns1-05.azure-dns.com.
    ns2-05.azure-dns.net.
    ns3-05.azure-dns.org.
    ns4-05.azure-dns.info.
    ```

    - Set up your domain to use these name servers.

4. Check if everything is set up correctly and your domain is managed by Azure name servers. Run:

   ```bash
   host -t ns $DNS_DOMAIN
   ```

A successful response returns the list of the name servers you fetched from Azure.

  </details>

</div>

## Generate the TLS certificate

<div tabs name="tls-certificate-generation" group="use-your-own-domain">
  <details>
  <summary label="GKE">
  GKE
  </summary>

Get the TLS certificate:

1. Export the certificate issuer email and the cluster domain as environment variables:

    ```bash
    export CERT_ISSUER_EMAIL={YOUR_EMAIL}
    export DOMAIN="$CLUSTER_NAME.$(echo $DNS_NAME | sed 's/\.$//')"
    ```

2. Create a folder for certificates. Run:

    ```bash
    mkdir letsencrypt
    ```

3. Create a new service account and assign it to the **dns.admin** role. Run these commands:

    ```bash
    gcloud iam service-accounts create dnsmanager --display-name "dnsmanager" --project "$GCP_PROJECT"
    gcloud projects add-iam-policy-binding $GCP_PROJECT \
        --member serviceAccount:dnsmanager@$GCP_PROJECT.iam.gserviceaccount.com --role roles/dns.admin
    ```

    > **NOTE**: You don't have to create a new DNS manager service account (SA) every time you deploy a cluster. Instead, you can use an existing SA that has the **dns.admin** assigned.

4. Generate an access key for this account in the `letsencrypt` folder. Run:

    ```bash
    gcloud iam service-accounts keys create ./letsencrypt/key.json --iam-account dnsmanager@$GCP_PROJECT.iam.gserviceaccount.com
    ```

    > **NOTE**: The number of keys you can generate for a single service account is limited. Reuse the existing keys instead of generating a new key for every cluster.

5. Run the Certbot Docker image with the `letsencrypt` folder mounted. Certbot uses the key to apply DNS challenge for the certificate request and stores the TLS certificates in that folder. Run:

    ```bash
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

6. Export the certificate and the key as environment variables. Run these commands:

    ```bash
    export TLS_CERT=$(cat ./letsencrypt/live/$DOMAIN/fullchain.pem | base64 | sed 's/ /\\ /g' | tr -d '\n');
    export TLS_KEY=$(cat ./letsencrypt/live/$DOMAIN/privkey.pem | base64 | sed 's/ /\\ /g' | tr -d '\n')
    ```

  </details>
  <details>
  <summary label="AKS">
  AKS
  </summary>

Get the TLS certificate:

>**NOTE:** Azure DNS is not yet supported by Certbot so you must perform manual verification.

1. Create a folder for certificates. Run:

    ```bash
    mkdir letsencrypt
    ```

2. Export your email address as an environment variable:

    ```bash
    export YOUR_EMAIL={YOUR_EMAIL}
    ```

3. To get the certificate, run the Certbot Docker image with the `letsencrypt` folder mounted. Certbot stores the TLS certificates in that folder.

    ```bash
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

    ```bash
    Please deploy a DNS TXT record under the name
    _acme-challenge.rc2-test.kyma.online with the following value:

    # TXT_VALUE

    Before continuing, verify the record is deployed.
    ```

    Copy the `TXT_VALUE`.

4. Open a new terminal and export these environment variables:

    ```bash
    export DNS_DOMAIN={YOUR_DOMAIN} # example.com
    export SUB_DOMAIN={YOUR_SUBDOMAIN} # cluster (in this case the full name of your cluster is cluster.example.com)
    export RS_GROUP={YOUR_RESOURCE_GROUP_NAME}
    ```

5. Export the `TXT_VALUE`.

    ```bash
    export TXT_VALUE={YOUR_TXT_VALUE}
    ```

    To modify the TXT record for your domain, run:

    ```bash
    az network dns record-set txt delete -n "_acme-challenge.$SUB_DOMAIN" -g $RS_GROUP -z $DNS_DOMAIN --yes
    az network dns record-set txt create -n "_acme-challenge.$SUB_DOMAIN" -g $RS_GROUP -z $DNS_DOMAIN --ttl 60 > /dev/null
    az network dns record-set txt add-record -n "_acme-challenge.$SUB_DOMAIN" -g $RS_GROUP -z $DNS_DOMAIN --value $TXT_VALUE
    ```

6. Go back to the first console, wait for about 2 minutes and press **Enter**.

7. Export the certificate and the key as environment variables. Run these commands:

    ```bash
    export TLS_CERT=$(cat ./letsencrypt/live/$SUB_DOMAIN.$DNS_DOMAIN/fullchain.pem | base64 | sed 's/ /\\ /g' | tr -d '\n')
    export TLS_KEY=$(cat ./letsencrypt/live/$SUB_DOMAIN.$DNS_DOMAIN/privkey.pem | base64 | sed 's/ /\\ /g' | tr -d '\n')
    ```

  </details>

</div>

## Prepare the cluster

<div tabs name="prepare-cluster" group="use-your-own-domain">
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
   >**NOTE**: Kyma offers the production profile. Pass the flag `-t` to Kyma CLI with `n1-standard-8` or `c2-standard-8` if you want to use it.

4. Configure kubectl to use your new cluster:

    ```bash
    gcloud container clusters get-credentials $CLUSTER_NAME --zone $GCP_ZONE --project $GCP_PROJECT
    ```

5. Add your account as the cluster administrator:

    ```bash
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)
    ```
    >**TIP:** See a [sample ConfigMap](./assets/owndomain-overrides.yaml) for reference.

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

3. Create a [service principle](https://docs.microsoft.com/en-us/azure/aks/kubernetes-service-principal#manually-create-a-service-principal) on Azure. Create a TOML file with the Azure Client ID, Client Secret, Subscription ID and Tenant ID:

    ```toml
    CLIENT_ID = {YOUR_CLIENT_ID}
    CLIENT_SECRET = {YOUR_CLIENT_SECRET}
    SUBSCRIPTION_ID = {YOUR_SUBSCRIPTION_ID}
    TENANT_ID = {YOUR_TENANT_ID}
    ```

4. Create an AKS cluster:

    ```bash
    kyma provision aks -c {YOUR_CREDENTIALS_FILE_PATH} -n $CLUSTER_NAME -p $RS_GROUP -l $REGION
    ```
   >**NOTE**: Kyma offers a production profile. Pass the flag `-t` to Kyma CLI with `Standard_F8s_v2` or `Standard_D8_v3` if you want to use it.

5. Add additional privileges to be able to access readiness probes endpoints on your AKS cluster.

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/$KYMA_VERSION/installation/resources/azure-crb-for-healthz.yaml
    ```
    >**CAUTION:** If you define your own Kubernetes jobs on the AKS cluster, follow the [troubleshooting guide](/components/service-mesh/#troubleshooting-kubernetes-jobs-fail-on-aks) to avoid jobs running endlessly on AKS deployments of Kyma.

  </details>

</div>

## Install Kyma

   >**NOTE**: If you want to use the Kyma production profile, see the following documents before you go to the next step:
      >* [Istio production profile](/components/service-mesh/#configuration-service-mesh-production-profile)
      >* [OAuth2 server production profile](/components/security/#configuration-o-auth2-server-profiles)

1. Install Kyma using Kyma CLI:

    ```bash
    kyma install -s $KYMA_VERSION --domain $DOMAIN --tlsCert $TLS_CERT --tlsKey $TLS_KEY
    ```

## Configure DNS for the cluster load balancer

<div tabs name="configure-dns" group="use-your-own-domain">
  <details>
  <summary label="GKE">
  GKE
  </summary>

To add DNS entries, run these commands:

```bash
export EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

export APISERVER_PUBLIC_IP=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

gcloud dns --project=$GCP_PROJECT record-sets transaction start --zone=$DNS_ZONE

gcloud dns --project=$GCP_PROJECT record-sets transaction add $EXTERNAL_PUBLIC_IP --name=\*.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE

gcloud dns --project=$GCP_PROJECT record-sets transaction add $APISERVER_PUBLIC_IP --name=\apiserver.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE

gcloud dns --project=$GCP_PROJECT record-sets transaction execute --zone=$DNS_ZONE
```

  </details>
  <details>
  <summary label="AKS">
  AKS
  </summary>

To add DNS entries, run these commands:

```bash
export EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

export APISERVER_PUBLIC_IP=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

az network dns record-set a create -g $RS_GROUP -z $DNS_DOMAIN -n \*.$SUB_DOMAIN --ttl 60
az network dns record-set a add-record -g $RS_GROUP -z $DNS_DOMAIN -n \*.$SUB_DOMAIN -a $EXTERNAL_PUBLIC_IP

az network dns record-set a create -g $RS_GROUP -z $DNS_DOMAIN -n apiserver.$SUB_DOMAIN --ttl 60
az network dns record-set a add-record -g $RS_GROUP -z $DNS_DOMAIN -n apiserver.$SUB_DOMAIN -a $APISERVER_PUBLIC_IP
```

  </details>

</div>

### Access the cluster

1. To open the cluster's Console in your default browser, run:

    ```bash
    kyma console
    ```

2. To log in to your cluster's Console UI, use the default `admin` static user. Click **Login with Email** and sign in with the **admin@kyma.cx** email address. Use the password printed after the installation. To get the password manually, you can also run:

    ```bash
    kubectl get secret admin-user -n kyma-system -o jsonpath="{.data.password}" | base64 --decode
    ```

If you need to use Helm to manage your Kubernetes resources, read the [additional configuration](#installation-use-helm) document.
