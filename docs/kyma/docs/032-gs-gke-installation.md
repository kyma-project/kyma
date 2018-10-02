---
title: Cluster Kyma installation
type: Getting Started
---

## Prepare own kyma-installer image

You can skip this step if you want to install image created from unmodified sources. 

1. Checkout kyma-project and enter root folder.
2. Build image that will include current installation and resources charts and is based on current installer image (you can find it in the installation/resources/installer.yaml file)
    ```
    docker build -t kyma-installer:latest -f kyma-installer/kyma.Dockerfile . --build-arg INSTALLER_VERSION=63484523
    ```
3. Push image to docker hub:
    ```
    docker tag kyma-installer:latest [YOUR_DOCKER_LOGIN]/kyma-installer:latest


## DNS setup

1. Get some domain and delegate it to google. You can get free domain from freenom.com site. I reserved kyma.ga domain.

2. Create DNS managed zone in your google project. You can create if from console.cloud.google.com. Navigate to Network Services, Cloud DNS and select Create Zone. Command line version:    
    ```
    gcloud dns --project=$PROJECT managed-zones create kymaga-zone --description= --dns-name=kyma.ga.
    ```
3. Delegate your domain to google name servers. List of nameservers you can get from zone details. In my case:
    - ns-cloud-b1.googledomains.com.
    - ns-cloud-b2.googledomains.com.
    - ns-cloud-b3.googledomains.com.
    - ns-cloud-b4.googledomains.com.

At freenom.com you can do that going to Services -> Main Domains - > Manage domain. Then select management tools -> Nameservers and enter values retrieved from google.

## Get TLS certificate

Create folder for certificates. Create new service account, assign it to role dns.admin and generate access key for this account in the lets encrypt folder. Then run certbot docker image with letsencrypt folder mounted. Certbot will use the key to apply DNS challenge for the certificate request and will store TLS certificates in that folder.
```
mkdir letsencrypt

gcloud iam service-accounts create dnsmanager --display-name "dnsmanager"

gcloud projects add-iam-policy-binding $PROJECT \
    --member serviceAccount:dnsmanager@$PROJECT.iam.gserviceaccount.com --role roles/dns.admin

gcloud iam service-accounts keys create ./letsencrypt/key.json --iam-account dnsmanager@$PROJECT.iam.gserviceaccount.com

export DOMAIN=demo.kyma.ga

sudo docker run -it --name certbot --rm \
    -v "$(pwd)/letsencrypt:/etc/letsencrypt" \
    certbot/dns-google \
    certonly \
    -m YOUR_EMAIL_HERE --agree-tos --no-eff-email \
    --dns-google \
    --dns-google-credentials /etc/letsencrypt/key.json \
    --server https://acme-v02.api.letsencrypt.org/directory \
    -d "*.$DOMAIN"

export TLS_CERT=$(cat ./letsencrypt/live/$DOMAIN/cert.pem | base64)
export TLS_KEY=$(cat ./letsencrypt/live/$DOMAIN/privkey.pem | base64)
```

## Prepare GKE cluster


1. Set environment variables for your Google project identifier (you have to change it) and set some cluster name (you can keep demo name).
    ```
    export CLUSTER_NAME=demo2
    export PROJECT=sap-hybris-sf-playground
    ```
2. Create cluster in europe-west1 region.
    ```
    gcloud beta container --project "$PROJECT" clusters create "$CLUSTER_NAME" --zone "europe-west1-b" --cluster-version "1.10.7-gke.2" --machine-type "n1-standard-2" 
    ```

## Prepare kyma installation.yaml

```
cat installation/resources/tiller.yaml installation/resources/installer.yaml <(echo "---") installation/resources/installer-config-cluster.yaml.tpl  <(echo "---") installation/resources/installer-cr.yaml.tpl | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" |sed -e "s/__.*__//g" >my-kyma.yaml

gcloud container clusters get-credentials $CLUSTER_NAME --zone europe-west1-b --project $PROJECT

kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)

kubectl apply -f my-kyma.yaml
```

Check if all pods are running (tiller, kyma-installer)
```
 kubectl get pods --all-namespaces
```

and finaly start kyma installation
```
kubectl label installation/kyma-installation action=install
```
Watch installation progress by:
```
kubectl logs -n kyma-installer [kyma-installer-pod] -f
```
or
```
kubectl get pods --all-namespaces -w
```

## Configure DNS for LB

```
export EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

export REMOTE_ENV_IP=$(kubectl get service -n kyma-system core-nginx-ingress-controller -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

gcloud dns --project=$PROJECT record-sets transaction start --zone=kymaga-zone

gcloud dns --project=$PROJECT record-sets transaction add $EXTERNAL_PUBLIC_IP --name=\*.$DOMAIN. --ttl=60 --type=A --zone=kymaga-zone

gcloud dns --project=$PROJECT record-sets transaction add $REMOTE_ENV_IP --name=\gateway.$DOMAIN. --ttl=60 --type=A --zone=kymaga-zone

gcloud dns --project=$PROJECT record-sets transaction execute --zone=kymaga-zone

```