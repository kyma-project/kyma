---
title: Install Kyma on an AKS cluster
type: Installation
---

This Installation guide shows developers how to quickly deploy Kyma on an [Azure Kubernetes Service](https://azure.microsoft.com/services/kubernetes-service/) (AKS) cluster. Kyma installs on a cluster using a proprietary installer based on a Kubernetes operator.

## Prerequisites
- A domain for your AKS cluster
- [Microsoft Azure](https://azure.microsoft.com)
- [Docker](https://www.docker.com/)
- [Docker Hub](https://hub.docker.com/) account
- [az](https://docs.microsoft.com/pl-pl/cli/azure/install-azure-cli)
- set the environment variables

### Environment variables

Set the following environment variables:
* To set mandatory variables, run:
  ```
  export RS_GROUP={YOUR_RESOURCE_GROUP_NAME}
  export CLUSTER_NAME={YOUR_CLUSTER_NAME}
  export REGION={YOUR_REGION} #westeurope
  ```

* If you want to install Kyma from release,  choose a [release](https://github.com/kyma-project/kyma/releases/) you want to use and export it as an environment variable. Run:
  ```
  export KYMA_RELEASE_VERSION={KYMA_RELEASE_VERSION}
  ```
  >**NOTE:** Installation from release is supported since version 0.7.0. 
  If you want to install Kyma from sources, do not export `KYMA_RELEASE_VERSION`.

* If you use a custom domain, set also these variables:
  ```
  export DNS_DOMAIN={YOUR_DOMAIN} # example.com
  export SUB_DOMAIN={YOUR_SUBDOMAIN} # cluster (in this case the full name of your cluster is cluster.example.com)
  ```

Create a resource group that will contain all your resources:
```
az group create --name $RS_GROUP --location $REGION
```

>**NOTE:** If you don't own a domain which you can use or you don't want to assign a domain to a cluster, see the [document](#installation-install-kyma-on-an-aks-cluster-with-wildcard-dns) which shows you how to create a cluster-based playground environment using a wildcard DNS provided by xip.io. 

## DNS setup

Delegate the management of your domain to Azure DNS. Follow these steps:


1. Create a DNS-managed zone in your Azure subscription. Run:

    ```
    az network dns zone create -g $RS_GROUP -n $DNS_DOMAIN
    ```

    Alternatively, create it through the Azure UI. In the **Networking** section, go to **All services**, click **DNS zones**, and select **Add**.

2. Delegate your domain to Azure name servers.

    - Get the list of the name servers from the zone details. This is a sample list:
      ```
      ns1-05.azure-dns.com.
      ns2-05.azure-dns.net.
      ns3-05.azure-dns.org.
      ns4-05.azure-dns.info.
      ```

    - Set up your domain to use these name servers.

3. Check if everything is set up correctly and your domain is managed by Azure name servers. Run:
    ```
    host -t ns $DNS_DOMAIN
    ```
    A successful response returns the list of the name servers you fetched from Azure.

## Get the TLS certificate

>**NOTE:** Azure DNS is not yet supported by Certbot so you must perform a manual verification.

1. Create a folder for certificates. Run:
    ```
    mkdir letsencrypt
    ```
2. Run the Certbot Docker image with the `letsencrypt` folder mounted. Certbot stores the TLS certificates in that folder. Export your email address:
    ```
    export YOUR_EMAIL={YOUR_EMAIL}
    ```
    To obtain a certificate, run:
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
    
3. Open a new console and set the environment variables from the [Environment variables](#installation-install-kyma-on-an-aks-cluster-environment-variables) section. Export the `TXT_VALUE`.
    
    ```
    export TXT_VALUE={YOUR_TXT_VALUE}
    ```
    To modify TXT record for your domain, run:
    ```
    az network dns record-set txt delete -n "_acme-challenge.$SUB_DOMAIN" -g $RS_GROUP -z $DNS_DOMAIN --yes
    az network dns record-set txt create -n "_acme-challenge.$SUB_DOMAIN" -g $RS_GROUP -z $DNS_DOMAIN --ttl 60 > /dev/null
    az network dns record-set txt add-record -n "_acme-challenge.$SUB_DOMAIN" -g $RS_GROUP -z $DNS_DOMAIN --value $TXT_VALUE
    ``` 
4. Go back to the first console, wait 2 minutes and press enter. 

5. Export the certificate and key as environment variables. Run these commands:

    ```
    export TLS_CERT=$(cat ./letsencrypt/live/$SUB_DOMAIN.$DNS_DOMAIN/fullchain.pem | base64 | sed 's/ /\\ /g')
    export TLS_KEY=$(cat ./letsencrypt/live/$SUB_DOMAIN.$DNS_DOMAIN/privkey.pem | base64 | sed 's/ /\\ /g')
    ```

## Prepare the AKS cluster

1. Create an AKS cluster. Run:
    ```
    az aks create \
      --resource-group $RS_GROUP \
      --name $CLUSTER_NAME \
      --node-vm-size "Standard_DS2_v2" \
      --kubernetes-version 1.10.9 \
      --enable-addons "monitoring,http_application_routing" \
      --generate-ssh-keys
    ```
2. To configure kubectl to use your new cluster, run:
    ```
    az aks get-credentials --resource-group $RS_GROUP --name $CLUSTER_NAME
    ```

3. Install Tiller and add additional privileges to be able to access readiness probes endpoints on your AKS cluster.
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

## Prepare the installation configuration file

### Using the latest GitHub release

1. Download the `kyma-config-cluster.yaml` and `kyma-installer-cluster.yaml` files from the latest release. Run:
   ```
   wget https://github.com/kyma-project/kyma/releases/download/$KYMA_RELEASE_VERSION/kyma-config-cluster.yaml
   wget https://github.com/kyma-project/kyma/releases/download/$KYMA_RELEASE_VERSION/kyma-installer-cluster.yaml
   ```

2. Update the file with the values from your environment variables. Merge files from step 3 to one `my-kyma.yaml` file. Run:
    ```
    cat kyma-installer-cluster.yaml <(echo -e "\n---") kyma-config-cluster.yaml | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g"|sed -e "s/__.*__//g"  >my-kyma.yaml
    ```

3. The output of this operation is the `my_kyma.yaml` file. Use it to deploy Kyma on your AKS cluster.


### Using your own image

1. Build an image that is based on the current Installer image and includes the current installation and resources charts. Run:

    ```
    docker build -t kyma-installer:latest -f tools/kyma-installer/kyma.Dockerfile .
    ```

2. Push the image to your Docker Hub:
    ```
    docker tag kyma-installer:latest {YOUR_DOCKER_LOGIN}/kyma-installer:latest
    ```
    ```
    docker push {YOUR_DOCKER_LOGIN}/kyma-installer:latest
    ```

3. Prepare the deployment file:

    ```
    cat installation/resources/installer.yaml <(echo -e "\n---") installation/resources/installer-config-cluster.yaml.tpl  <(echo -e "\n---") installation/resources/installer-cr-cluster.yaml.tpl | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" |sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__.*__//g" > my-kyma.yaml
    ```

4. In the `my-kyma.yaml` file created in the previous step, change the image URL to `{YOUR_DOCKER_LOGIN}/kyma-installer:latest`.
    ```
    kind: Deployment
    metadata:
      name: kyma-installer
      namespace: kyma-installer
      ......
        image: eu.gcr.io/kyma-project/develop/installer:30bf314d
    ```
    Use the modified file to deploy Kyma on your AKS cluster.

## Deploy Kyma

1. Deploy Kyma using the `my-kyma` custom configuration file you created. Run:
    ```
    kubectl apply -f my-kyma.yaml
    ```
    >**NOTE:** In case you receive the `Error from server (MethodNotAllowed)` error, run the command again before going to step 2.
    
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


## Configure DNS for the cluster load balancer

Run these commands:

```
export EXTERNAL_PUBLIC_IP=$(kubectl get service -n istio-system istio-ingressgateway -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

export REMOTE_ENV_IP=$(kubectl get service -n kyma-system application-connector-nginx-ingress-controller -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

az network dns record-set a create -g $RS_GROUP -z $DNS_DOMAIN -n \*.$SUB_DOMAIN --ttl 60
az network dns record-set a add-record -g $RS_GROUP -z $DNS_DOMAIN -n \*.$SUB_DOMAIN -a $EXTERNAL_PUBLIC_IP

az network dns record-set a create -g $RS_GROUP -z $DNS_DOMAIN -n gateway.$SUB_DOMAIN --ttl 60
az network dns record-set a add-record -g $RS_GROUP -z $DNS_DOMAIN -n gateway.$SUB_DOMAIN -a $REMOTE_ENV_IP
```

Access your cluster under this address:
```
https://console.$SUB_DOMAIN.$DNS_DOMAIN
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
    cat kyma-installer-cluster.yaml <(echo -e "\n---") cat kyma-config-cluster.yaml | sed -e "s/__DOMAIN__/$SUB_DOMAIN.$DNS_DOMAIN/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g"  | sed -e "s/__TLS_CERT__/$TLS_CERT/g" | sed -e "s/__TLS_KEY__/$TLS_KEY/g" | sed -e "s/__REMOTE_ENV_CA__/$AC_CRT/g" | sed -e "s/__REMOTE_ENV_CA_KEY__/$AC_KEY/g" |sed -e "s/__.*__//g"  >my-kyma.yaml
    ```
