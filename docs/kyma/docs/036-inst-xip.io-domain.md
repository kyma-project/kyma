---
title: Install Kyma on a GKE cluster with wildcard DNS
type: Installation
---

If you want to try Kyma in a cluster environment without assigning the cluster to a domain you own, you can use [`xip.io`](http://xip.io/) which provides a wildcard DNS for any IP address. Such
a scenario requires using a self-signed TLS certificate.

This solution is not suitable for a production environment but makes for a great playground which allows you to get to know the product better.

## Prerequisites

The prerequisites match these listed in [this](#installation-install-kyma-on-a-gke-cluster) document. However, you don't need to prepare a domain for your cluster as it is replaced by a wildcard DNS provided by [`xip.io`](http://xip.io/).

>**NOTE:** This feature requires Kyma version 0.6 or higher.

## Installation

The installation process follows the steps outlined in the [Install Kyma on a GKE cluster](#installation-install-kyma-on-a-gke-cluster) document. Follow [this](#installation-install-kyma-on-a-gke-cluster-prepare-the-gke-cluster) section to prepare your cluster.

In addition to exporting the desired cluster name as an environment variable, make sure to export your GCP project name. Run:
```
export PROJECT={YOUR_GCP_PROJECT_NAME}
```

When you install Kyma with the wildcard DNS, you can use one of two approaches to allocating the required IP addresses for your cluster:
- Dynamic IP allocation - can be used with [Knative](#installation-installation-with-knative) eventing and serverless, but disables the Application Connector. 
- Manual IP allocation - cannot be used with [Knative](#installation-installation-with-knative) eventing and serverless, but leaves the Application Connector functional. 

Follow the respective instructions to deploy a cluster Kyma cluster with wildcard DNS which uses the IP allocation approach of your choice.

### Dynamic IP allocation

1. Use this command to prepare a configuration file that deploys Kyma with [`xip.io`](http://xip.io/) providing a wildcard DNS:
```
(cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster-xip-io.yaml.tpl) | sed -e "s/__.*__//g" > my-kyma.yaml
```
>**NOTE:** Using this approach disables the Application Connector. 

2. Follow [these](#installation-install-kyma-on-a-gke-cluster-deploy-kyma) instructions to install Kyma using the configuration file you prepared.

### Manual IP allocation

1. Get public IP addresses for the load balancer of the GKE cluster to which you deploy Kyma and for the load balancer of the Application Connector.

  - Export the `PUBLIC_IP_ADDRESS_NAME` and the `APP_CONNECTOR_IP_ADDRESS_NAME` environment variables. This defines the names of the reserved public IP addresses in your GCP project. Run:
    ```
    export PUBLIC_IP_ADDRESS_NAME={GCP_COMPLIANT_PUBLIC_IP_ADDRESS_NAME}
    export APP_CONNECTOR_IP_ADDRESS_NAME={GCP_COMPLIANT_APP_CONNECTOR_IP_ADDRESS_NAME}
    ```
    >**NOTE:** The name you set for the reserved public IP address must start with a lowercase letter followed by up to 62 lowercase letters, numbers, or hyphens, and cannot end with a hyphen.

  - Run these commands to reserve public IP addresses for the load balancer of your cluster and the load balancer of the Application Connector.
    ```
    gcloud beta compute --project=$PROJECT addresses create $PUBLIC_IP_ADDRESS_NAME --region=europe-west1 --network-tier=PREMIUM
    gcloud beta compute --project=$PROJECT addresses create $APP_CONNECTOR_IP_ADDRESS_NAME --region=europe-west1 --network-tier=PREMIUM
    ```
    >**NOTE:** The region in which you reserve IP addresses must match the region of your GKE cluster.

  - Set the reserved IP addresses as `EXTERNAL_PUBLIC_IP` and `CONNCETOR_IP` environment variables. Run:
    ```
    export EXTERNAL_PUBLIC_IP=$(gcloud compute addresses list --project=$PROJECT --filter="name=$PUBLIC_IP_ADDRESS_NAME" --format="value(address)")
    export CONNECTOR_IP=$(gcloud compute addresses list --project=$PROJECT --filter="name=$APP_CONNECTOR_IP_ADDRESS_NAME" --format="value(address)")
    ```

2. Use this command to prepare a configuration file that deploys Kyma with [`xip.io`](http://xip.io/) providing a wildcard DNS:
  ```
(cat installation/resources/installer.yaml ; echo "\n---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "\n---" ; cat installation/resources/installer-cr-cluster-xip-io.yaml.tpl) | sed -e "s/__EXTERNAL_PUBLIC_IP__/$EXTERNAL_PUBLIC_IP/g" | sed -e "s/__REMOTE_ENV_IP__/$CONNECTOR_IP/g" | sed -e "s/__APPLICATION_CONNECTOR_DOMAIN__/$CONNECTOR_IP.xip.io/g" | sed -e "s/__SKIP_SSL_VERIFY__/true/g" | sed -e "s/__.*__//g" > my-kyma.yaml
  ```
3. Follow [these](#installation-install-kyma-on-a-gke-cluster-deploy-kyma) instructions to install Kyma using the configuration file you prepared.  


### Add the xip.io self-signed certificate to your OS trusted certificates

After the installation, add the custom Kyma [`xip.io`](http://xip.io/) self-signed certificate to the trusted certificates of your OS. For MacOS run:
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
