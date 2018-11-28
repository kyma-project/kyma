---
title: Services and Plans
type: Details
---

## Service description

The `gcp-broker-provider` service provides the following plan names and descriptions:

| Plan Name | Description |
|-----------|-------------|
| `Default` | Installs the Google Cloud Platform (GCP) Service Broker in a default configuration. |

## Provisioning

>**NOTE:** You can provision only one instance of the GCP Service Broker Provider in each Namespace.

To add GCP Broker Provider to your Namespace, you must prepare a service account and a 
JSON access key. 

Follow these steps to create a Kubernetes Secret which contains a JSON access key:
1. Open https://console.cloud.google.com/ and select your project.
2. Go to **IAM & admin** -> **Service accounts**.
3. Click **Create service account**.
4. Assign `Project Owner` role.
5. Click **Create key** and choose `JSON` as key type.
6. Save file to a known location.
7. Create a secret from the JSON file:

```
kubectl create secret generic gcp-broker-data --from-file=sa-key={filename} --from-literal=project-name=kyma-project --namespace {namespace}
```

Please note that you have to create a secret in every namespace where the GCP Broker Provider class is provisioned.

In the Service Catalog view click **Google Cloud Platform Service Broker Provider**.
Provisioning of this class adds GCP Service Broker classes to the Service Catalog in a given namespace.

The service account key created by user is used to 
generate service account keys used by brokers installed in different namespaces.
The generated service account key has a **roles/servicebroker.operator** role and is 
used during provisioning/deprovisioning/binding/unbinding actions.

![](assets/gcp-broker-key-management.svg)

Flow of the provisioning process is described by following diagram

![GCP Broker Provisioning](assets/gcp-broker-provisioning.svg)

1. User triggers provision action
2. During the provisioning process new service account and access key are created in **Google Cloud Platform**
3. After the provisioning process is over the post install job is triggered.

## Binding

Binding to this Service Class is disabled.

## Deprovisioning

Flow of the deprovisioning process is described by following diagram


![GCP Broker Deprovisioning](assets/gcp-broker-deprovisioning.svg)

1. User triggers deprovision action
2. If secret is present access keys are removed from **Google Cloud Platform**


