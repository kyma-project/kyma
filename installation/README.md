# Kyma Installer

## Overview

The Kyma installer is a proprietary solution based on the Kubernetes operator. See the [Local Kyma installation](../docs/kyma/docs/031-gs-local-installation.md) document for basic information on how to use the installer and run Kyma locally. Read further sections to learn how to install Kyma on a cluster.

## Cluster installation

**NOTE:** Currently we assume that public IPs and DNS records for istio ingress and remote environments gateway exists before Kyma installation. Because AWS does not support assigning static IPs during creating ELB, installer does not support clusters on AWS. We are working to change that behaviour soon.

**NOTE:** Before you start the installation, check that you have access to a running cluster with the minimum Kubernetes version 1.10.

To install Kyma, you need the following data:

- IP address for Kyma Ingress
- IP address for Remote Environments Ingress [Read more](../docs/application-connector/docs/006-architecture-ingress-gateway.md)
- Domain name (for example `kyma.example.com`)
  - `gateway.kyma.example.com` points to Remote Environments Ingress IP address
  - `*.kyma.example.com` points to Kyma Ingress IP address
- Wildcard TLS certificate for your cluster domain that you can generate with Let's Encrypt
- Certificate for Remote Environments [Read more](../docs/application-connector/docs/001-overview-application-connector.md)

### Prerequisites

Kubernetes API Server is configured in the following manner:

```
"apiServerConfig": {
    "--enable-admission-plugins": "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,DefaultStorageClass,ResourceQuota,PodPreset",
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

### Steps

1. Create the `kyma-installer` Namespace.

Run the following command:

```
kubectl create ns kyma-installer
```

2. Fill in the `installer-config.yaml.tpl` template.

The Kyma installation process requires installation data specified in the `installer-config.yaml` file. Copy the `installer-config.yaml.tpl` template, rename it to `installer-config.yaml`, and fill in these placeholder values:

- `__TLS_CERT__` for the TLS certificate
- `__TLS_KEY__` for the TLS certificate key
- `__REMOTE_ENV_CA__` for the Remote Environments CA [Read more](../docs/application-connector/docs/001-overview-application-connector.md)
- `__REMOTE_ENV_CA_KEY__` for the Remote Environments CA key [Read more](../docs/application-connector/docs/001-overview-application-connector.md)
- `__EXTERNAL_IP_ADDRESS__` for the IP address for Kyma Ingress
- `__DOMAIN__` for the domain name, for example `kyma.example.com`
- `__REMOTE_ENV_IP__` for the IP address for Remote Environments Ingress. [Read more](../docs/application-connector/docs/006-architecture-ingress-gateway.md)
- `__K8S_APISERVER_URL__` for the API server's URL
- `__K8S_APISERVER_CA__` for your API Server CA
- `__ADMIN_GROUP__` for the additional admin group. This value is optional.
- `__ENABLE_ETCD_BACKUP_OPERATOR__` for enabling or disabling the `etcd` backup operator. Enter `true` or `false`.
- `__ETCD_BACKUP_ABS_CONTAINER_NAME__` for the Azure Blob Storage name for `etcd` backups. You can leave the value blank when the backup operator is disabled.

**NOTE:** Currently `etcd` backup feature is in development, set `__ENABLE_ETCD_BACKUP_OPERATOR__` to `false` and leave `__ETCD_BACKUP_ABS_CONTAINER_NAME__` blank.

When you fill in all required placeholder values, run the following command to provide the cluster with the installation data:

```
kubectl apply -f installer-config.yaml
```

3. Deploy the `Installer` component.

To deploy the `Installer` component on your cluster, run the following command:

```
kubectl apply -f installation/resources/installer.yaml -n kyma-installer
```

4. Trigger the installation.

To trigger the installation of Kyma, you need a Custom Resource file. Copy the `installer-cr.yaml.tpl` file, rename it to `installer-cr.yaml`, and fill in these placeholder values:

- `__VERSION__` for the version number of Kyma to install. When manually installing Kyma on a cluster, specify any valid [SemVer](https://semver.org/) notation string, for example `0.0.1`
- `__URL__` for the URL to the Kyma `tar.gz` package to install. For example, for current master branch of Kyam, the url will be `https://github.com/kyma-project/kyma/archive/master.tar.gz`

Once the file is ready, run the following command to trigger the installation:

```
kubectl apply -f installer-cr.yaml
```

To check the progress of the installation process, verify the Custom Resource:

```
kubectl get installation kyma-installation -o yaml
```

Successful installation ends setting `status.state` to `Installed` and `status.description` to `Kyma installed`

To troubleshoot installation, start with reviewing logs from `Installer` component:

```
kubectl logs -n kyma-installer $(kubectl get pods --all-namespaces -l name=kyma-installer --no-headers -o jsonpath='{.items[*].metadata.name}')
```