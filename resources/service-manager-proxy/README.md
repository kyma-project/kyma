# service-broker-proxy-k8s

Service Manager Broker Proxy for Kubernetes

Changes added to the original upstream chart:
1. In deployment.yaml file such annotation was added:
    ```yaml
    annotations:
            sidecar.istio.io/inject: "false" 
    ```
2. In __helpers.tpl_ file the `{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}` was replaced with `{{- .Release.Name | trunc 63 | trimSuffix "-" -}}`. As a result we have shorten and more readable names for Deployments, Secrets, ConfigMaps etc.

## Introduction

This helm chart bootstraps the Service Broker Proxy for Kubernetes.

## Prerequisites

* `tiller` is installed and configured in the Kubernetes cluster.
* `helm` is installed and configured.
* `service-catalog` is installed and configured in the Kubernetes cluster.
* The cluster is registered in the Service Manager as a platform.

You can register the cluster in Service Manager by executing the following command:
```sh
smctl register-platform <cluster-name> kubernetes
```
**Note:** Store the returned credentials in a safe place as you will not be able to fetch them again from Service Manager.

You can get `smctl` tool from https://github.com/Peripli/service-manager-cli.

## Installation

### Add Peripli helm repo
```bash
helm repo add peripli 'https://peripli.github.io'
```

### Instal Service Manager agent

```bash
# using helm v2.x.x
helm install peripli/service-broker-proxy-k8s \
  --name service-broker-proxy \
  --namespace service-broker-proxy \
  --set image.tag=<VERSION> \
  --set config.sm.url=<SM_URL> \
  --set sm.user=<USER> \
  --set sm.password=<PASSWORD>
```

```bash
# using helm v3.x.x
kubectl create namespace service-broker-proxy
helm install service-broker-proxy peripli/service-broker-proxy-k8s \
  --namespace service-broker-proxy \
  --set image.tag=<VERSION> \
  --set config.sm.url=<SM_URL> \
  --set sm.user=<USER> \
  --set sm.password=<PASSWORD>
```

**Note:** Make sure you substitute &lt;SM_URL&gt; with the Service Manager url, &lt;USER&gt; and &lt;PASSWORD&gt; with the credentials returned from cluster registration in Service Manager (see above).
Substitute \<VERSION> with the required version as listed on [Releases](https://github.com/Peripli/service-broker-proxy-k8s/releases). It is recommended to use the latest release.

To use your own images you can set `image.repository`, `image.tag` and `image.pullPolicy` to the helm install command. In case your image is pulled from a private repository, you can use
`image.pullsecret` to name a secret containing the credentials.

## Configuration

The following table lists some of the configurable parameters of the service broker proxy for K8S chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`image.repository` | image repository |`quay.io/service-manager/sb-proxy-k8s`
`image.tag` | tag of image | `master`
`image.pullsecret` | name of the secret containing pull secrets |
`config.sm.url` | service manager url | `http://service-manager.dev.cfdev.sh`
`sm.user` | username for service manager | `admin`
`sm.password` | password for service manager | `admin`
`securityContext` | Custom [security context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) for server containers | `{}`
