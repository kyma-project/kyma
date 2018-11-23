---
title: Services and Plans
type: Details
---

## Service description

This service is named `gcp-broker-provider` with the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Default` | Installs the GCP Service Broker in a default configuration |

>**Note:** There can be only one instance of the GCP Service Broker Provider in every namespace.

## Provision

To add GCP Broker Provider to your namespace you first need to prepare a service account and a 
JSON access key. 

### Prerequisites

To create a Kubernetes Secret entry containing JSON access key perform the following steps:
1. Open https://console.cloud.google.com/ and select your project.
2. Go to **IAM & admin** -> **Service accounts**.
3. Click **Create service account**.
4. Assign `Project Owner` role.
5. Click **Create key** and choose `JSON` as key type.
6. Save file to a known location.
7. Create a secret from the JSON file:

   ```kubectl create secret generic gcp-broker-data --from-file=sa-key={filename} --from-literal=project-name=kyma-project --namespace {namespace}```

Please note that you have to create a secret in every namespace where the GCP Broker Provider class is provisioned.

### Installation

In the Service Catalog view click **Google Cloud Platform Service Broker Provider**.
Provisioning of this class adds GCP Service Broker classes to the Service Catalog.

![GCP Broker Classes](assets/gcp-broker-classes.png)


### Details

The service account key created by user is used to 
generate service account keys used by brokers installed in different namespaces.
The generate service account key has a **roles/servicebroker.operator** role and is 
used during provisioning/deprovisioning/binding/unbinding actions.


![](assets/gcp-broker-key-management.svg)

### Credentials

>**Note:** Binding to this service class is disabled

