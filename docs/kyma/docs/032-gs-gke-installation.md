---
title: Install Kyma on a GKE cluster
type: Getting Started
---

This Getting Started guide shows developers how to quickly deploy Kyma on a [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/) (GKE) cluster. Kyma installs on a cluster using a proprietary installer based on a Kubernetes operator.

## Prerequisites

- A domain for your GKE cluster
- [Google Cloud Platform](https://console.cloud.google.com/) (GCP) project
- [Docker](https://www.docker.com/)
- [Docker Hub](https://hub.docker.com/) account
- [gcloud](https://cloud.google.com/sdk/gcloud/)

## DNS setup

Delegate the management of your domain to Google Cloud DNS. Follow these steps:

1. Export the domain name, project name and DNS zone name as environment variables. Run the commands listed below:

    ```
    export DOMAIN={YOUR_SUBDOMAIN}
    export DNS_NAME={YOUR_DOMAIN}.
    export PROJECT={YOUR_GOOGLE_PROJECT}
    export DNS_ZONE={YOUR_DNS_ZONE}
    ```

2. Create a DNS-managed zone in your Google project. Run:

    ```
    gcloud dns --project=$PROJECT managed-zones create $DNS_ZONE --description= --dns-name=$DNS_NAME
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

## Get the TLS certificate

1. Create a folder for certificates. Run:
    ```
    mkdir letsencrypt
    ```
2. Create a new service account and assign it to the `dns.admin` role. Run these commands:
    ```
    gcloud iam service-accounts create dnsmanager --display-name "dnsmanager"
    ```
    ```
    gcloud projects add-iam-policy-binding $PROJECT \
        --member serviceAccount:dnsmanager@$PROJECT.iam.gserviceaccount.com --role roles/dns.admin
    ```

3. Generate an access key for this account in the `letsencrypt` folder. Run:
    ```
    gcloud iam service-accounts keys create ./letsencrypt/key.json --iam-account dnsmanager@$PROJECT.iam.gserviceaccount.com
    ```
4. Run the Certbot Docker image with the `letsencrypt` folder mounted. Certbot uses the key to apply DNS challenge for the certificate request and stores the TLS certificates in that folder. Run:
    ```
    sudo docker run -it --name certbot --rm \
        -v "$(pwd)/letsencrypt:/etc/letsencrypt" \
        certbot/dns-google \
        certonly \
        -m YOUR_EMAIL_HERE --agree-tos --no-eff-email \
        --dns-google \
        --dns-google-credentials /etc/letsencrypt/key.json \
        --server https://acme-v02.api.letsencrypt.org/directory \
        -d "*.$DOMAIN"
    ```

5. Export the certificate and key as environment variables. Run these commands:

    ```
    export TLS_CERT=$(cat ./letsencrypt/live/$DOMAIN/fullchain.pem | base64)
    ```
    ```
    export TLS_KEY=$(cat ./letsencrypt/live/$DOMAIN/privkey.pem | base64)
    ```


## Prepare the GKE cluster

1. Select a name for your cluster and set it as an environment variable. Run:
    ```
    export CLUSTER_NAME={CLUSTER_NAME_YOU_WANT}
    ```

2. Create a cluster in the `europe-west1` region. Run:
    ```
    gcloud beta container --project "$PROJECT" clusters \
    create "$CLUSTER_NAME" --zone "europe-west1-b" \
    --cluster-version "1.10.7-gke.6" --machine-type "n1-standard-2" \
    --addons HorizontalPodAutoscaling,HttpLoadBalancing,KubernetesDashboard
    ```

3. Install Tiller on your GKE cluster. Run:

    ```
    kubectl apply -f installation/resources/tiller.yaml
    ```

## Prepare the installation configuration file

### Using the latest GitHub release

1. Download the `kyma-config-cluster` file bundled with the latest Kyma [release](https://github.com/kyma-project/kyma/releases/).

2. Update the file with the values from your environment variables. Run:
    ```
    cat kyma-config-cluster.yaml | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g"|sed -e "s/__.*__//g"  >my-kyma.yaml
    ```

3. The output of this operation is the `my_kyma.yaml` file. Use it to deploy Kyma on your GKE cluster.


### Using your own image

1. Checkout [kyma-project](https://github.com/kyma-project/kyma) and enter the root folder.

2. Build an image that is based on the current installer image and includes the current installation and resources charts. Run:

    ```
    docker build -t kyma-installer:latest -f kyma-installer/kyma.Dockerfile . --build-arg INSTALLER_VERSION=63484523
    ```

3. Push the image to your Docker Hub:
    ```
    docker tag kyma-installer:latest [YOUR_DOCKER_LOGIN]/kyma-installer:latest
    ```
    ```
    docker push [YOUR_DOCKER_LOGIN]/kyma-installer:latest
    ```

4. Prepare the deployment file:

    ```
    cat installation/resources/installer.yaml <(echo -e "\n---") installation/resources/installer-config-cluster.yaml.tpl  <(echo -e "\n---") installation/resources/installer-cr-cluster.yaml.tpl | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

5. The output of this operation is the `my_kyma.yaml` file. Modify it to fetch the proper image with the changes you made ([YOUR_DOCKER_LOGIN]/kyma-installer:latest). Use the modified file to deploy Kyma on your GKE cluster.


## Deploy Kyma

1. Configure kubectl to use your new cluster. Run:  add yourself as the cluster admin, and deploy Kyma installer with your configuration.
    ```
    gcloud container clusters get-credentials $CLUSTER_NAME --zone europe-west1-b --project $PROJECT
    ```
2. Add your account as the cluster administrator:
    ```
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)
    ```
3. Deploy Kyma using the `my-kyma` custom configuration file you created. Run:
    ```
    kubectl apply -f my-kyma.yaml
    ```
4. Check if the Pods of Tiller and the Kyma installer are running:
    ```
    kubectl get pods --all-namespaces
    ```

5. Start Kyma installation:
    ```
    kubectl label installation/kyma-installation action=install
    ```

6. To watch the installation progress, run:
    ```
    kubectl get pods --all-namespaces -w
    ```


## Configure DNS for the cluster load balancer

Run these commands:

```
export EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

export REMOTE_ENV_IP=$(kubectl get service -n kyma-system application-connector-nginx-ingress-controller -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

gcloud dns --project=$PROJECT record-sets transaction start --zone=$DNS_ZONE

gcloud dns --project=$PROJECT record-sets transaction add $EXTERNAL_PUBLIC_IP --name=\*.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE

gcloud dns --project=$PROJECT record-sets transaction add $REMOTE_ENV_IP --name=\gateway.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE

gcloud dns --project=$PROJECT record-sets transaction execute --zone=$DNS_ZONE

```

## Prepare your Kyma deployment for production use

To use the cluster in a production environment, it is recommended you configure a new server-side certificate for the Application Connector and replace the placeholder certificate it installs with.
If you don't generate a new certificate, the system uses the placeholder certificate. As a result, the security of your implementation is compromised.

Follow this steps to configure a new, more secure certificate suitable for production use.

1. Generate a new certificate and key. Run:

    ```
    openssl req -new -newkey rsa:4096 -nodes -keyout ca.key -out ca.csr -subj "/C=PL/ST=N/L=GLIWICE/O=SAP Hybris/OU=Kyma/CN=wormhole.kyma.cx"

    openssl x509 -req -sha256 -days 365 -in ca.csr -signkey ca.key -out ca.pem
    ```

2. Export the certificate and key to environment variables:

    ```
    export AC_CRT=$(cat ./ca.pem | base64 | base64)
    export AC_KEY=$(cat ./ca.key | base64 | base64)

    ```

3. Prepare installation file with the following command:

    ```
    cat kyma-config-cluster.yaml | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__REMOTE_ENV_CA__/$AC_CRT/g" | sed -e "s/__REMOTE_ENV_CA_KEY__/$AC_KEY/g" |sed -e "s/__.*__//g"  >my-kyma.yaml
    ```
