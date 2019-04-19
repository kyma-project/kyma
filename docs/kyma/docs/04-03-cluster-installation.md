---
title: Install Kyma on a cluster
type: Installation
---

This Installation guide shows developers how to quickly deploy Kyma on a cluster. Kyma is installed on a cluster using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/). By default, Kyma is installed on a cluster with a wildcard DNS provided by [xip.io](http://xip.io). Alternatively, you can provide your own domain for the cluster.

Follow these installation guides to install Kyma on a cluster depending on the supported cloud providers:
<div tabs>
  <details>
  <summary>
  GKE
  </summary>

This Installation guide shows developers how to quickly deploy Kyma on a [Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/) (GKE) cluster.

## Prerequisites
- [Google Cloud Platform](https://console.cloud.google.com/) (GCP) project with Kubernetes Engine API enabled
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 1.12.0
- [Docker](https://www.docker.com/)
- [Docker Hub](https://hub.docker.com/) account
- [gcloud](https://cloud.google.com/sdk/gcloud/)
- [wget](https://www.gnu.org/software/wget/)
- A domain for your GKE cluster (optional)

>**TIP:** Get a free domain for your cluster using services like [freenom.com](https://www.freenom.com) or similar.

## Prepare the GKE cluster

1. Select a name for your cluster. Set the cluster name and the name of your GCP project as environment variables. Run:
    ```
    export CLUSTER_NAME={CLUSTER_NAME_YOU_WANT}
    export PROJECT={YOUR_GCP_PROJECT}
    ```

2. Create a cluster in the `europe-west1` region. Run:
    ```
    gcloud container --project "$PROJECT" clusters \
    create "$CLUSTER_NAME" --zone "europe-west1-b" \
    --cluster-version "1.12" --machine-type "n1-standard-4" \
    --addons HorizontalPodAutoscaling,HttpLoadBalancing,KubernetesDashboard
    ```

3. Install Tiller on your GKE cluster. Run:

    ```
    kubectl apply -f https://raw.githubusercontent.com/kyma-project/kyma/{RELEASE_TAG}/installation/resources/tiller.yaml
    ```

4. Add your account as the cluster administrator:
    ```
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=$(gcloud config get-value account)
    ```

## DNS setup and TLS certificate generation (optional)

>**NOTE:** Execute instructions from this section only if you want to use your own domain. Otherwise, proceed to **Prepare the installation configuration file** section.

### Delegate the management of your domain to Google Cloud DNS

Follow these steps:

1. Export the domain name, project name, and DNS zone name as environment variables. Run the commands listed below:

    ```
    export DOMAIN={YOUR_SUBDOMAIN}
    export DNS_NAME={YOUR_DOMAIN}.
    export PROJECT={YOUR_GOOGLE_PROJECT_ID}
    export DNS_ZONE={YOUR_DNS_ZONE}
    ```
    Example:
    ```
    export DOMAIN=my.kyma-demo.ga
    export DNS_NAME=kyma-demo.ga.
    export PROJECT=kyma-demo-235208
    export DNS_ZONE=myzone
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

### Get the TLS certificate

1. Create a folder for certificates. Run:
    ```
    mkdir letsencrypt
    ```
2. Create a new service account and assign it to the `dns.admin` role. Run these commands:
    ```
    gcloud iam service-accounts create dnsmanager --display-name "dnsmanager" --project "$PROJECT"
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
    docker run -it --name certbot --rm \
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
    export TLS_CERT=$(cat ./letsencrypt/live/$DOMAIN/fullchain.pem | base64 | sed 's/ /\\ /g')
    export TLS_KEY=$(cat ./letsencrypt/live/$DOMAIN/privkey.pem | base64 | sed 's/ /\\ /g')
    ```

## Prepare the installation configuration file

### Using the latest GitHub release

>**NOTE:** You can use Kyma version 0.8 or higher.

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

    - Run this command if you use the `xip.io` default domain:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__DOMAIN__/$DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    
    > **NOTE:** If you deploy Kyma with GKE version 1.12.6-gke.X and above, follow these steps to prepare the deployment file. 
        
    - Run this command if you use the xip.io default domain:
        
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    
    - Run this command if you use your own domain:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__DOMAIN__/$DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    
5. The output of this operation is the `my_kyma.yaml` file. Use it to deploy Kyma on your GKE cluster.

### Using your own image

1. Checkout [kyma-project](https://github.com/kyma-project/kyma) and enter the root folder.

2. Build an image that is based on the current Installer image and includes the current installation and resources charts. Run:

    ```
    docker build -t kyma-installer:latest -f tools/kyma-installer/kyma.Dockerfile .
    ```

3. Push the image to your Docker Hub:
    ```
    docker tag kyma-installer:latest {YOUR_DOCKER_LOGIN}/kyma-installer:latest
    docker push {YOUR_DOCKER_LOGIN}/kyma-installer:latest
    ```

4. Prepare the deployment file:

    - Run this command if you use the `xip.io` default domain:
    ```
    (cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    (cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    > **NOTE:** If you deploy Kyma with GKE version 1.12.6-gke.X and above, follow these steps to prepare the deployment file. 
        
    - Run this command if you use the xip.io default domain:
    ```
    (cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    (cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__DOMAIN__/$DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    
5. The output of this operation is the `my_kyma.yaml` file. Modify it to fetch the proper image with the changes you made ([YOUR_DOCKER_LOGIN]/kyma-installer:latest). Use the modified file to deploy Kyma on your GKE cluster.

## Deploy Kyma

1. Configure kubectl to use your new cluster. Run:
    ```
    gcloud container clusters get-credentials $CLUSTER_NAME --zone europe-west1-b --project $PROJECT
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

## Add the xip.io self-signed certificate to your OS trusted certificates

>**NOTE:** Skip this section if you use your own domain.

After the installation, add the custom Kyma [`xip.io`](http://xip.io/) self-signed certificate to the trusted certificates of your OS. For MacOS, run:
  ```
  tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
  && kubectl get configmap  net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.tlsCrt}'  | base64 --decode > $tmpfile \
  && sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
  && rm $tmpfile
  ```

## Configure DNS for the cluster load balancer (optional)

>**NOTE:** Execute instructions from this section only if you want to use your own domain.

1. Export the domain of your cluster and DNS zone as environment variables. Run:
    ```
    export DOMAIN=$(kubectl get cm net-global-overrides -n kyma-installer -o jsonpath='{.data.global\.ingress\.domainName}')
    export DNS_ZONE={YOUR_DNS_ZONE}
    ```

2. To add DNS entries, run these commands:
    ```
    export EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

    export APISERVER_PUBLIC_IP=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

    export REMOTE_ENV_IP=$(kubectl get service -n kyma-system application-connector-ingress-nginx-ingress-controller -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

    gcloud dns --project=$PROJECT record-sets transaction start --zone=$DNS_ZONE

    gcloud dns --project=$PROJECT record-sets transaction add $EXTERNAL_PUBLIC_IP --name=\*.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE

    gcloud dns --project=$PROJECT record-sets transaction add $REMOTE_ENV_IP --name=\gateway.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE

    gcloud dns --project=$PROJECT record-sets transaction add $APISERVER_PUBLIC_IP --name=\apiserver.$DOMAIN. --ttl=60 --type=A --zone=$DNS_ZONE

    gcloud dns --project=$PROJECT record-sets transaction execute --zone=$DNS_ZONE
    ```
  </details>
  <details>
  <summary>
  AKS
  </summary>

This Installation guide shows developers how to quickly deploy Kyma on an [Azure Kubernetes Service](https://azure.microsoft.com/services/kubernetes-service/) (AKS) cluster.

## Prerequisites
- [Microsoft Azure](https://azure.microsoft.com)
- [Kubernetes](https://kubernetes.io/) 1.12
- Tiller 2.10.0 or higher
- [Docker](https://www.docker.com/)
- [Docker Hub](https://hub.docker.com/) account
- [az](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)
- A domain for your AKS cluster (optional)

>**TIP:** Get a free domain for your cluster using services like [freenom.com](https://www.freenom.com) or similar.

## Prepare the AKS cluster

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
    
## DNS setup and TLS certificate generation (optional)

>**NOTE:** Execute instructions from this section only if you want to use your own domain. Otherwise, proceed to **Prepare the installation configuration file** section.

### Delegate the management of your domain to Azure DNS

Follow these steps:

1. Export the domain name, the sub-domain, and the resource group name as environment variables. Run these commands:

    ```
    export DNS_DOMAIN={YOUR_DOMAIN} # example.com
    export SUB_DOMAIN={YOUR_SUBDOMAIN} # cluster (in this case the full name of your cluster is cluster.example.com)
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

## Prepare the installation configuration file

### Using the latest GitHub release

>**NOTE:** You can use Kyma version 0.8 or higher.

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

    - Run this command if you use the `xip.io` default domain:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    
    > **NOTE:** If you deploy Kyma with Kubernetes version 1.14 and above, follow these steps to prepare the deployment file. 
        
    - Run this command if you use the xip.io default domain:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    
5. The output of this operation is the `my_kyma.yaml` file. Use it to deploy Kyma on your GKE cluster.


### Using your own image

1. Checkout [kyma-project](https://github.com/kyma-project/kyma) and enter the root folder.

2. Build an image that is based on the current Installer image and includes the current installation and resources charts. Run:

    ```
    docker build -t kyma-installer:latest -f tools/kyma-installer/kyma.Dockerfile .
    ```

3. Push the image to your Docker Hub:
    ```
    docker tag kyma-installer:latest {YOUR_DOCKER_LOGIN}/kyma-installer:latest
    docker push {YOUR_DOCKER_LOGIN}/kyma-installer:latest
    ```

4. Prepare the deployment file:

    - Run this command if you use the `xip.io` default domain:
    ```
    (cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    (cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    > **NOTE:** If you deploy Kyma with Kubernetes version 1.14 and above, follow these steps to prepare the deployment file. 
    - Run this command if you use the xip.io default domain:
    ```
    (cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

    - Run this command if you use your own domain:
    ```
    (cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster.yaml.tpl) | sed -e "s/__PROMTAIL_CONFIG_NAME__/promtail-k8s-1-14.yaml/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```
    
5. The output of this operation is the `my_kyma.yaml` file. Modify it to fetch the proper image with the changes you made ([YOUR_DOCKER_LOGIN]/kyma-installer:latest). Use the modified file to deploy Kyma on your GKE cluster.

## Deploy Kyma

1. Deploy Kyma using the `my-kyma` custom configuration file you created. Run:
    ```
    kubectl apply -f my-kyma.yaml
    ```
    >**NOTE:** If you get the `Error from server (MethodNotAllowed)` error, run the command again before proceeding to the next step.

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

## Add the xip.io self-signed certificate to your OS trusted certificates

>**NOTE:** Skip this section if you use your own domain.

After the installation, add the custom Kyma [`xip.io`](http://xip.io/) self-signed certificate to the trusted certificates of your OS. For MacOS, run:
```
tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
&& kubectl get configmap cluster-certificate-overrides -n kyma-installer -o jsonpath='{.data.global\.tlsCrt}' | base64 --decode > $tmpfile \
&& sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
&& rm $tmpfile
```

## Configure DNS for the cluster load balancer (optional)

>**NOTE:** Execute instructions from this section only if you want to use your own domain.

Run these commands:

```
export EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

export REMOTE_ENV_IP=$(kubectl get service -n kyma-system application-connector-ingress-nginx-ingress-controller -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

export APISERVER_PUBLIC_IP=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

az network dns record-set a create -g $RS_GROUP -z $DNS_DOMAIN -n \*.$SUB_DOMAIN --ttl 60
az network dns record-set a add-record -g $RS_GROUP -z $DNS_DOMAIN -n \*.$SUB_DOMAIN -a $EXTERNAL_PUBLIC_IP

az network dns record-set a create -g $RS_GROUP -z $DNS_DOMAIN -n gateway.$SUB_DOMAIN --ttl 60
az network dns record-set a add-record -g $RS_GROUP -z $DNS_DOMAIN -n gateway.$SUB_DOMAIN -a $REMOTE_ENV_IP

az network dns record-set a create -g $RS_GROUP -z $DNS_DOMAIN -n apiserver.$SUB_DOMAIN --ttl 60
az network dns record-set a add-record -g $RS_GROUP -z $DNS_DOMAIN -n apiserver.$SUB_DOMAIN -a $APISERVER_PUBLIC_IP
```
  </details>
</div>

## Access Tiller (optional)

If you need to use Helm, you must establish a secure connection with Tiller by saving the cluster's client certificate, key, and Certificate Authority (CA) to [Helm Home](https://helm.sh/docs/glossary/#helm-home-helm-home). 

Additionally, you must add the `--tls` flag to every Helm command you run.

>**NOTE:** Read [this](#details-tls-in-tiller) document to learn more about TLS in Tiller.

Run these commands to save the client certificate, key, and CA to [Helm Home](https://helm.sh/docs/glossary/#helm-home-helm-home):

```bash
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem";
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem";
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem";
```

## Access the cluster

1. To get the address of the cluster's Console, check the name of the Console's virtual service. The name of this virtual service corresponds to the Console URL. To get the virtual service name, run:

```
kubectl get virtualservice core-console -n kyma-system
```

2. Access your cluster under this address:

```
https://{VIRTUAL_SERVICE_NAME}
```

>**NOTE:** To log in to your cluster, use the default `admin` static user. To learn how to get the login details for this user, see [this](#installation-install-kyma-locally-access-the-kyma-console) document.
