---
title: Install Kyma on an AKS cluster with wildcard DNS
type: Installation
---

If you want to try Kyma in a cluster environment without assigning the cluster to a domain you own, you can use [`xip.io`](http://xip.io/) which provides a wildcard DNS for any IP address. Such
a scenario requires using a self-signed TLS certificate.

This solution is not suitable for a production environment but makes for a great playground which allows you to get to know the product better.

## Prerequisites

The prerequisites match these listed in [this](#installation-install-kyma-on-an-aks-cluster) document. However, you don't need to prepare a domain for your cluster as it is replaced by a wildcard DNS provided by [`xip.io`](http://xip.io/).

>**NOTE:** This feature requires Kyma version 0.6 or higher.

## Installation

The installation process follows the steps outlined in the [Install Kyma on an AKS cluster](#installation-install-kyma-on-an-aks-cluster) document. Set [environment variables](#installation-install-kyma-on-an-aks-cluster-environment-variables) and follow [this](#installation-install-kyma-on-an-aks-cluster-prepare-the-aks-cluster) section to prepare your cluster.

When you install Kyma with the wildcard DNS, you can use one of two approaches to allocating the required IP addresses for your cluster:
- Dynamic IP allocation - can be used with [Knative](#installation-installation-with-knative) eventing and serverless, but disables the Application Connector. 
- Manual IP allocation - cannot be used with [Knative](#installation-installation-with-knative) eventing and serverless, but leaves the Application Connector functional. 

Follow the respective instructions to deploy a cluster Kyma cluster with wildcard DNS which uses the IP allocation approach of your choice.

### Dynamic IP allocation

Use this command to prepare a configuration file that deploys Kyma with [`xip.io`](http://xip.io/) providing a wildcard DNS:
```
(cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster-xip-io.yaml.tpl) | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" | sed -e "s/__.*__//g" > my-kyma.yaml
```
>**NOTE:** Using this approach disables the Application Connector. 

### Manual IP allocation

1. Export your Kubernetes cluster resource group. This group is different from the one you provided during the cluster creation. It is automatically created by AKS. 
   Set the same set of environment variables as during the [cluster initialization](#installation-install-kyma-on-an-aks-cluster-environment-variables).
   Run:
   ```
   export CLUSTER_RS_GROUP=MC_"$RS_GROUP"_"$CLUSTER_NAME"_"$REGION"
   ```
2. Get public IP addresses for the load balancer of the AKS cluster to which you deploy Kyma and for the load balancer of the Application Connector.

  - Export the `PUBLIC_IP_ADDRESS_NAME` and the `APP_CONNECTOR_IP_ADDRESS_NAME` environment variables. This defines the names of the reserved public IP addresses in your Azure subscription. Run:
    ```
    export PUBLIC_IP_ADDRESS_NAME={AZURE_COMPLIANT_PUBLIC_IP_ADDRESS_NAME}
    export APP_CONNECTOR_IP_ADDRESS_NAME={AZURE_COMPLIANT_APP_CONNECTOR_IP_ADDRESS_NAME}
    ```
    >**NOTE:** The name you set for the reserved public IP address must start with a lowercase letter followed by up to 62 lowercase letters, numbers, or hyphens, and cannot end with a hyphen.

  - Run these commands to reserve public IP addresses for the load balancer of your cluster and the load balancer of the Application Connector.
    ```
    az network public-ip create --name $PUBLIC_IP_ADDRESS_NAME --resource-group $CLUSTER_RS_GROUP --allocation-method Static
    az network public-ip create --name $APP_CONNECTOR_IP_ADDRESS_NAME --resource-group $CLUSTER_RS_GROUP --allocation-method Static
    ```
    >**NOTE:** The resource group in which you reserve IP addresses must match the resource group of your AKS cluster.

  - Set the reserved IP addresses as `EXTERNAL_PUBLIC_IP` and `CONNCETOR_IP` environment variables. Run:
    ```
    export EXTERNAL_PUBLIC_IP=$(az network public-ip list -g $CLUSTER_RS_GROUP --query "[?name=='$PUBLIC_IP_ADDRESS_NAME'].ipAddress" -otsv)
    export CONNECTOR_IP=$(az network public-ip list -g $CLUSTER_RS_GROUP --query "[?name=='$APP_CONNECTOR_IP_ADDRESS_NAME'].ipAddress" -otsv)
    ```
3. Use this command to prepare a configuration file that deploys Kyma with [`xip.io`](http://xip.io/) providing a wildcard DNS:
  ```
(cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster-xip-io.yaml.tpl) | sed -e "s/__EXTERNAL_PUBLIC_IP__/$EXTERNAL_PUBLIC_IP/g" | sed -e "s/__REMOTE_ENV_IP__/$CONNECTOR_IP/g" | sed -e "s/__APPLICATION_CONNECTOR_DOMAIN__/$CONNECTOR_IP.xip.io/g" | sed -e "s/__SKIP_SSL_VERIFY__/true/g" | sed -e "s/__PROXY_EXCLUDE_IP_RANGES__/10.0.0.1/g" |  sed -e "s/__.*__//g" > my-kyma.yaml
  ```

### Kyma installation

You can either choose the pre-build image of the Kyma Installer or build your own.

* To build your own image:
  1. Build an image that is based on the current Installer image and includes the current installation and resources charts. Run:
     ```
     docker build -t kyma-installer:latest -f tools/kyma-installer/kyma.Dockerfile . --build-arg INSTALLER_VERSION=63484523
     ```
  2. Push the image to your Docker Hub:
     ```
     docker tag kyma-installer:latest [YOUR_DOCKER_LOGIN]/kyma-installer:latest
     docker push [YOUR_DOCKER_LOGIN]/kyma-installer:latest
     ```
* To use a prebuild image, go to [this](https://github.com/kyma-project/kyma/releases/) page and check the version of the latest release. Your URL looks as follows:
```eu.gcr.io/kyma-project/kyma-installer:{latest version}```

In the `my-kyma.yaml` file, change the image URL to the value taken from the previous step.
```
kind: Deployment
metadata:
  name: kyma-installer
  namespace: kyma-installer
......
        image: eu.gcr.io/kyma-project/develop/installer:30bf314d
```
Follow [these](#installation-install-kyma-on-an-aks-cluster-deploy-kyma) instructions to install Kyma using the configuration file you prepared.

### Add the xip.io self-signed certificate to your OS trusted certificates

After the installation, add the custom Kyma [`xip.io`](http://xip.io/) self-signed certificate to the trusted certificates of your OS. For MacOS, run:
```
tmpfile=$(mktemp /tmp/temp-cert.XXXXXX) \
&& kubectl get configmap cluster-certificate-overrides -n kyma-installer -o jsonpath='{.data.global\.tlsCrt}' | base64 --decode > $tmpfile \
&& sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $tmpfile \
&& rm $tmpfile
```

## Access the cluster

To access your cluster, use the wildcard DNS provided by [`xip.io`](http://xip.io/) as the domain of the cluster. To get this information, run:
```
kubectl get cm installation-config-overrides -n kyma-installer -o jsonpath='{.data.global\.domainName}'
```
A successful response returns the cluster domain following this format:
```
{WILDCARD_DNS}.xip.io
```
Access your cluster under this address:
```
https://console.{WILDCARD_DNS}.xip.io
```

>**NOTE:** To log in to your cluster, use the default `admin` static user. To learn how to get the login details for this user, see [this](#installation-install-kyma-locally-from-the-release-access-the-kyma-console) document. 
