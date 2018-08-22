---
title: Cluster Kyma installation
type: Getting Started
---

## Overview

This Getting Started guide shows developers how to quickly deploy Kyma on a cluster. Kyma installs on a cluster using a proprietary installer based on a [Kubernetes operator](https://coreos.com/operators/). The document provides prerequisites, instructions on how to install Kyma on a cluster and verify the deployment, as well as the troubleshooting tips.

>**NOTE:** To learn how to install Kyma on a [Gardener](https://github.com/gardener/gardener) cluster, see the **Install Kyma on Gardener** document.

## Prerequisites

The cluster on which you install Kyma must run Kubernetes version `1.10` or higher.

Prepare these items: 

- A domain name such as `kyma.example.com`.
- A wildcard TLS certificate for your cluster domain. Generate it with [**Let's Encrypt**](https://letsencrypt.org/).
- The certificate for Remote Environments.
- A static IP address for the Kyma Istio Ingress (public external IP). Create a DNS record `*.kyma.example.com` that points to Kyma Istio Ingress IP Address.
- A Static IP address for Remote Environments Ingress. Create a DNS record `gateway.kyma.example.com` that points to Remote Environments Ingress IP Address.

Some providers doen't allow to pre-allocate IP addresses, such as is the case with AWS which does not support static IP assignment during ELB creation. For such providers, you must complete the configuration after you install Kyma. See the **DNS configuration** section for more details. 

>**NOTE:** See the Application Connector documentation for more details on Remote Environments.

Configure the Kubernetes API Server following this template:

```
"apiServerConfig": {
    "--enable-admission-plugins": "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,DefaultStorageClass,ResourceQuota",
    "--runtime-config": "batch/v2alpha1=true,settings.k8s.io/v1alpha1=true,admissionregistration.k8s.io/v1alpha1=true",
    "--cors-allowed-origins": ".*",
    "--feature-gates": "ReadOnlyAPIDataVolumes=false",
    "--oidc-issuer-url": "https://dex.kyma.example.com",
    "--oidc-client-id": "kyma-client",
    "--oidc-username-claim": "email",
    "--oidc-groups-claim": "groups"
},
"kubeletConfig": {
    "--feature-gates": "ReadOnlyAPIDataVolumes=false",
    "--authentication-token-webhook": "true",
    "--authorization-mode": "Webhook"
}
```

## Installation

1. Create the `kyma-installer` Namespace.

Run the following command:

```
kubectl create ns kyma-installer
```

2. Enable Azure Broker (optional)

>**NOTE:** This instruction is applicable only when you install Kyma on a Gardener Cluster..

To enable the communication between the Shoot Cluster and the Azure Broker, you must pass Azure credentials to the cluster.

- Copy the `azure-broker-secret.yaml.tpl` template located in the `installation/resources` directory.
- Rename the file to `azure-broker-secret.yaml`.
- Replace the placeholder values with your Azure credentials.
- Run this command to pass the secret to the Shoot Cluster:
```
kubectl apply -f installation/resources/azure-broker-secret.yaml
```

3. Fill in the `installer-config.yaml.tpl` template.

The Kyma installation process requires installation data specified in the `installer-config.yaml` file. Copy the `installer-config.yaml.tpl` template, rename it to `installer-config.yaml`, and fill in these placeholder values:

- `__TLS_CERT__` for the TLS certificate
- `__TLS_KEY__` for the TLS certificate key
- `__REMOTE_ENV_CA__` for the Remote Environments CA
- `__REMOTE_ENV_CA_KEY__` for the Remote Environments CA key
- `__IS_LOCAL_INSTALLATION__` for controlling installation procedure. Set to `true` for local installation, otherwise cluster installation is assumed.
- `__DOMAIN__` for the domain name such as `kyma.example.com`
- `__EXTERNAL_PUBLIC_IP__` for the IP address of Kyma Istio Ingress (optional)
- `__REMOTE_ENV_IP__` for the IP address for Remote Environments Ingress (optional)
- `__K8S_APISERVER_URL__` for the API server's URL
- `__K8S_APISERVER_CA__` for your API Server CA
- `__ADMIN_GROUP__` for the additional admin group. This value is optional.
- `__ENABLE_ETCD_BACKUP_OPERATOR__` to enable or disable the `etcd` backup operator. Enter `true` or `false`.
- `__ETCD_BACKUP_ABS_CONTAINER_NAME__` for the Azure Blob Storage name of `etcd` backups. You can leave the value blank when the backup operator is disabled.

>**NOTE:** As the `etcd` backup feature is in development, set `__ENABLE_ETCD_BACKUP_OPERATOR__` to `false` and leave `__ETCD_BACKUP_ABS_CONTAINER_NAME__` blank.

When you fill in all required placeholder values, run the following command to provide the cluster with the installation data:

```
kubectl apply -f installer-config.yaml
```

4. Bind the default RBAC role.

Kyma installation requires increased permissions granted by the **cluster-admin** role. To bind the role to the default **ServiceAccount**, run the following command:

```
kubectl apply -f resources/cluster-prerequisites/default-sa-rbac-role.yaml
```

5. Deploy `tiller`.

To deploy the `tiller` component on your cluster, run the following command:

```
kubectl apply -f installation/resources/tiller.yaml
```

Wait until the `tiller` Pod is ready. Execute the following command to check that it is running:

```
kubectl get pods -n kube-system | grep tiller
```

6. Deploy the `Installer` component.

To deploy the `Installer` component on your cluster, run this command:

```
kubectl apply -f installation/resources/installer.yaml -n kyma-installer
```

7. Trigger the installation.

To trigger the installation of Kyma, you need a Custom Resource file. Duplicate the `installer-cr.yaml.tpl` file, rename it to `installer-cr.yaml`, and fill in these placeholder values:

- `__VERSION__` for the version number of Kyma to install. When manually installing Kyma on a cluster, specify any valid [SemVer](https://semver.org/) notation string. For example, `0.0.1`.
- `__URL__` for the URL to the Kyma `tar.gz` package to install. For example, for the `master` branch of Kyma, the address is `https://github.com/kyma-project/kyma/archive/master.tar.gz`.

>**NOTE:** Read the **Installation** document to learn more about the Custom Resource that controls the Kyma installer.

Once the file is ready, run this command to trigger the installation:

```
kubectl apply -f installer-cr.yaml
```
8. Verify the installation.

To check the progress of the installation process, verify the Custom Resource:

```
kubectl get installation kyma-installation -o yaml
```

A successful installation ends by setting `status.state` to `Installed` and `status.description` to `Kyma installed`.

## DNS configuration

If the cluster provider doesn't allow to pre-allocate IP addresses, the cluster gets the required details from the underlying cloud provider infrastructure. Get the allocated IP addresses and set up the DNS entries required for Kyma.

- List all Services and look for "LoadBalancer":
  ```
  kubectl get services --all-namespaces | grep LoadBalancer
  ```

- Find `istio-ingressgateway` in the `istio-system` Namespace. This entry specifies the IP address for the Kyma Ingress. Create a DNS entry `*.kyma.example.com` that points to this IP address.

- Find `core-nginx-ingress-controller` in the `kyma-system` Namespace. This entry specifies the IP address for the Remote Environments Ingress. Create a DNS entry `gateway.kyma.example.com` that points to this address.


## Troubleshooting

To troubleshoot the installation, start by reviewing logs of the `Installer` component:

```
kubectl logs -n kyma-installer $(kubectl get pods --all-namespaces -l name=kyma-installer --no-headers -o jsonpath='{.items[*].metadata.name}')
```
