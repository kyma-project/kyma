---
title: Install Kyma on an AKS cluster
type: Installation
---

This Installation guide shows developers how to quickly deploy Kyma on an [Azure Kubernetes Service](https://azure.microsoft.com/services/kubernetes-service/) (AKS) cluster. Kyma is installed on a cluster using a proprietary installer based on a Kubernetes operator.

By default, Kyma is installed on an AKS cluster with a wildcard DNS provided by [xip.io](http://xip.io). Alternatively, you can provide your own domain for the cluster.

## Prerequisites
- [Microsoft Azure](https://azure.microsoft.com)
- [Kubernetes](https://kubernetes.io/) 1.12
- Tiller 2.10.0 or higher
- [Docker](https://www.docker.com/)
- [Docker Hub](https://hub.docker.com/) account
- [az](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)
- A domain for your AKS cluster (optional)

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
    ```

## DNS setup and TLS certificate generation

>**NOTE:** Execute instructions from this section only if you want to use your own domain. Otherwise, proceed to [this](#installation-install-kyma-on-a-gke-cluster-prepare-the-installation-configuration-file) section.

### Delegate the management of your domain to Azure DNS

Follow these steps:

1. Export the domain name, and sub-domain as environment variables. Run the commands listed below:

    ```
    export DNS_DOMAIN={YOUR_DOMAIN} # example.com
    export SUB_DOMAIN={YOUR_SUBDOMAIN} # cluster (in this case the full name of your cluster is cluster.example.com)
      ```

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

5. The output of this operation is the `my_kyma.yaml` file. Modify it to fetch the proper image with the changes you made ([YOUR_DOCKER_LOGIN]/kyma-installer:latest). Use the modified file to deploy Kyma on your GKE cluster.

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

## Add the xip.io self-signed certificate to your OS trusted certificates

>**NOTE:** Skip this section if you use your own domain.

After the installation, add the custom Kyma [`xip.io`](http://xip.io/) self-signed certificate to the trusted certificates of your OS. For MacOS run:
```
tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
&& kubectl get configmap cluster-certificate-overrides -n kyma-installer -o jsonpath='{.data.global\.tlsCrt}' | base64 --decode > $tmpfile \
&& sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
&& rm $tmpfile
```

## Configure DNS for the cluster load balancer

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

## Access the cluster

Access your cluster under this address:

```
https://console.{SUB_DOMAIN}.{DNS_DOMAIN}
```

>**NOTE:** To log in to your cluster, use the default `admin` static user. To learn how to get the login details for this user, see [this](#installation-install-kyma-locally-from-the-release-access-the-kyma-console) document.
