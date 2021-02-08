---
title: Install Kyma with your own domain with cert-manager
type: Installation
---

This guide explains how to deploy Kyma on a cluster using your own domain and cert-manager.

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
   >**NOTE**: Kyma offers a production profile. Pass the flag `-t` to Kyma CLI with `Standard_F8s_v2` or `Standard_D8_v3` if you want to use it.

5. Add additional privileges to be able to access readiness probes endpoints on your AKS cluster.

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/$KYMA_VERSION/installation/resources/azure-crb-for-healthz.yaml
    ```
   >**CAUTION:** If you define your own Kubernetes jobs on the AKS cluster, follow the [troubleshooting guide](/components/service-mesh/#troubleshooting-kubernetes-jobs-fail-on-aks) to avoid jobs running endlessly on AKS deployments of Kyma.

  </details>

</div>

## Provide required data - TW PART

User has to create a ConfigMap 'kyma-ca-issuer' in default namespace with a [ClusterIssuer CR](https://cert-manager.io/docs/configuration/) inside. The name 'kyma-ca-issuer' and namespace 'default' is requires. Depending on which ClusterIssuer he chooses he needs to perform different steps.

If he provides a [SelfSigned](https://cert-manager.io/docs/configuration/selfsigned/) issuer, he needs to create an override `global.certificates.selfSigned=true`.

I think we could provide examples for GKE and AKS and add a note in which we explain that other issuers exist, but they were not tested or something like that (we cannot test all of them, the amount is too high). We can link to [the cert-manager documentation](https://cert-manager.io/docs/configuration/).

<div tabs name="tls-certificate-generation" group="use-your-own-domain">
  <details>
  <summary label="GKE">
  GKE
  </summary>

1. Create a new service account and assign it to the **dns.admin** role. Run these commands:

    ```bash
    gcloud iam service-accounts create dnsmanager --display-name "dnsmanager" --project "$GCP_PROJECT"
    gcloud projects add-iam-policy-binding $GCP_PROJECT \
        --member serviceAccount:dnsmanager@$GCP_PROJECT.iam.gserviceaccount.com --role roles/dns.admin
    ```

   > **NOTE**: You don't have to create a new DNS manager service account (SA) every time you deploy a cluster. Instead, you can use an existing SA that has the **dns.admin** assigned.

2. Generate an access key for this account. Run:

    ```bash
    gcloud iam service-accounts keys create ./key.json --iam-account dnsmanager@$GCP_PROJECT.iam.gserviceaccount.com
    ```

   > **NOTE**: The number of keys you can generate for a single service account is limited. Reuse the existing keys instead of generating a new key for every cluster.
   
3. Create a Secret with the generated access key. Run:

    ```bash
   kubectl create secret generic kyma-certs-service-account \
    --from-file=key.json -n cert-manager 
   ```

4. Create a ConfigMap with ClusterIssuer. See an example below:

    ```yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: kyma-ca-issuer
      namespace: default
    data:
      issuer: |
        apiVersion: cert-manager.io/v1
        kind: ClusterIssuer
        metadata:
          name: kyma-ca-issuer
        spec:
          acme:
            email: {EMAIL}
            server: https://acme-v02.api.letsencrypt.org/directory
            privateKeySecretRef:
              name: kyma-ca-cert
            solvers:
              - dns01:
                cloudDNS:
                  project: {PROJECT_NAME}
                  serviceAccountSecretRef:
                    name: kyma-certs-service-account
                    key: key.json
    ```

>**NOTE:** For more configuration options, see [the cert-manager ACME issuer documentation](https://cert-manager.io/docs/configuration/acme/).

  </details>
  <details>
  <summary label="AKS">
  AKS
  </summary>

//TODO

  </details>

</div>

## Install Kyma

>**NOTE**: If you want to use the Kyma production profile, see the following documents before you go to the next step:
>* [Istio production profile](/components/service-mesh/#configuration-service-mesh-production-profile)
>* [OAuth2 server production profile](/components/security/#configuration-o-auth2-server-profiles)

1. Install Kyma using Kyma CLI:

    ```bash
    kyma install -s $KYMA_VERSION --domain $DOMAIN
    ```
   
TW PART
>**NOTE**: Maybe add note here about adding the `global.certificates.selfSigned=true` override.

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

