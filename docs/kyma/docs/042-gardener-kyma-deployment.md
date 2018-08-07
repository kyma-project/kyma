---
title: Install Kyma on Gardener
type: Getting Started
---

## Overview

This Getting Started guide shows developers how to install Kyma on a [Gardener](https://github.com/gardener/gardener) cluster. The document provides the prerequisites and the instructions on how to install Kyma on a Gardener cluster.

## Prerequisites

- access to a Gardener project
- a domain that you control

>**NOTE:** This guide assumes you understand the [architecture of Gardener](https://github.com/gardener/documentation/wiki/Architecture) and the purpose of different cluster types.

## Configure your Gardener project

Configure these items in the project dashboard:
  - Add the Microsoft Azure Cloud secret is present in the **SECRETS** tab.
  - Add a service account to the **MEMBERS** tab.

>**NOTE:** The Kyma installer does not support clusters on AWS infrastructure as the provider does not support static IP assignment during ELB creation.

## Prepare the Shoot Cluster

1. Generate a TLS certificate

Shoot Clusters require a TLS certificate to run Kyma properly. To generate a certificate for your domain, use one of the available Certificate Authorities, such as [Let's Encrypt](https://letsencrypt.org/).

2. Configure the Shoot Cluster.

The deployment of a Gardener cluster requires data specified in the `shoot.yaml` file. Copy the [shoot.yaml.tpl](../../../installation/resources/gardener/shoot.yaml.tpl) template, rename it to `shoot.yaml`, and fill in the placeholder values:

- `__CLUSTER_NAME__` for the desired name of your cluster
- `__PROJECT_NAME__` for the name of your project
- `__PURPOSE__` for the purpose of the cluster (development, evaluation or production)
- `__AZURE_SECRET_NAME__` for the name of your Azure secret specified it the **SECRETS** tab in the Gardener dashboard
- `__TLS_CERT__` for the TLS certificate generated in the first step
- `__DOMAIN__` for the domain that you control

3. Deploy the Shoot cluster.

To deploy a Shoot Cluster, you must connect to the Garden Cluster first

- Go to the **MEMBERS** tab in the Gardener dashboard.
- Download the Garden Cluster's `kubeconfig` file from the **Service Account** section.
- Export the `KUBECONFIG` environment variable to connect to the Garden Cluster. Run:
 ```
export KUBECONFIG={PATH_TO_GARDEN_CLUSTER_KUBECONFIG_FILE}
```
 - Make sure you are connected to the right cluster:
```
kubectl cluster-info
```
 - Deploy the Shoot Cluster with the configuration file created using the `shoot.yaml.tpl` template. Run:
```
kubectl create -f {PATH_TO_THE_SHOOT_CLUSTER_CONFIG_FILE}
```
- Go to the **CLUSTERS** tab in the Gardener dashboard to track the deployment process. Wait until your cluster is ready.     

- To connect to the Shoot Cluster, choose it from the list in the Gardener dashboard and download the `kubeconfig` file.
Connect to the Shoot Cluster by exporting the `KUBECONFIG` environment variable:
```
export KUBECONFIG={PATH_TO_SHOOT_CLUSTER_KUBECONFIG_FILE}
```

4. Add the managed-standard Storage Class

To add the Storage Class, run this command:
```
kubectl apply -f installation/resources/gardener/managed-standard.yaml
```

## Install Kyma

To install Kyma on the Shoot Cluster, follow the instructions from the **Kyma Cluster installation** document.
