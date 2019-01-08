---
title: Install Kyma on a GKE cluster with wildcard DNS
type: Installation
---

If you want to try Kyma in a cluster environment without assigning the cluster to a domain you own, you can use [`xip.io`](http://xip.io/) which provides a wildcard DNS for any IP address. Such
a scenario requires using a self-signed TLS certificate.

This solution is not suitable for a production environment but makes for a great playground which allows you to get to know the product better.

## Prerequisites

The prerequisites match these listed in the **Install Kyma on a GKE cluster** document. However, you don't need to prepare a domain for your cluster as it is replaced by a wildcard DNS provided by [`xip.io`](http://xip.io/).

>**NOTE:** This feature requires Kyma version 0.6.

## Installation

The installation process follows the steps outlined in the **Install Kyma on a GKE cluster** document. Skip the DNS configuration of your Google project and start with the **Prepare the GKE cluster** section of the document.

1. In addition to exporting the desired cluster name as an environment variable, make sure to export your GCP project name. Run:
  ```
  export PROJECT={YOUR_GCP_PROJECT_NAME}
  ```

2. Get a public IP address for the load balancer of the GKE cluster to which you deploy Kyma.

  - Export the `IP_ADDRESS_NAME` environment variable. This defines the name of the reserved public IP address in your GCP project. Run:
    ```
    export IP_ADDRESS_NAME={GCP_COMPLIANT_IP_ADDRESS_NAME}
    ```
    >**NOTE:** The name you set for the reserved public IP address must start with a lowercase letter followed by up to 62 lowercase letters, numbers, or hyphens, and cannot end with a hyphen.

  - Run this command to reserve a public IP address for the load balancer of your cluster.
    ```  
    gcloud beta compute --project=$PROJECT addresses create $IP_ADDRESS_NAME --region=europe-west1 --network-tier=PREMIUM
    ```
    >**NOTE:** The region in which you reserve the IP address must match the region of your GKE cluster.

  - Set the reserved IP address as the `EXTERNAL_PUBLIC_IP` environment variable. Run:
    ```
    export EXTERNAL_PUBLIC_IP=$(gcloud compute addresses list --project=$PROJECT --filter="name=$IP_ADDRESS_NAME" --format="value(address)")
    ```

3. Use this command to prepare a configuration file that deploys Kyma with [`xip.io`](http://xip.io/) providing a wildcard DNS:
  ```
  (cat installation/resources/installer.yaml ; echo "---" ; cat installation/resources/installer-config-cluster.yaml.tpl ; echo "---" ; cat installation/resources/installer-cr-cluster-xip-io.yaml.tpl) | sed -e "s/__EXTERNAL_PUBLIC_IP__/$EXTERNAL_PUBLIC_IP/g" | sed -e "s/__.*__//g" > my-kyma.yaml
  ```

4. Follow the instructions from the **Deploy Kyma** section to install Kyma using the configuration file you prepared.

5. Add the custom Kyma [`xip.io`](http://xip.io/) self-signed certificate to the trusted certificates of your OS. For MacOS run:
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

>**NOTE:** To log in to your cluster, use the default `admin` static user. To learn how to get the login details for this user, see the **Access the Kyma console** section in the **Install Kyma locally from the release** document.
