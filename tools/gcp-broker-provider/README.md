# GCP Broker Provider

## Overview

This project provides the following scripts:
- `gcp-broker.sh` which installs or uninstalls the GCP Service Broker in a given Namespace
- `status-checker.sh` which checks if the GCP Service Broker is registered in the Service Catalog

## Prerequisites

To run the project scripts locally, these tools are required:
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [gcloud](https://cloud.google.com/sdk/)
* [sc](https://github.com/kyma-incubator/k8s-service-catalog)

To build the project's Docker image, only the latest version of the [Docker](https://www.docker.com/) tool is required.

## Usage

This section explains how to use the GCP Broker Provider scripts. Run all commands from the examples in the root directory of this project.

### Run `gcp-broker.sh` locally 

Before you proceed, ensure that the Secret with JSON access key and GCP project name exists in Namespace where the GCP Broker will be installed. 
Find the full guide on how to create JSON access key [here](https://cloud.google.com/video-intelligence/docs/common/auth#set_up_a_service_account). The assigned role must be the **project owner.** 

Run this command to create the required Secret:
```bash
kubectl create secret generic gcp-broker-data --from-file=sa-key={filename} --from-literal=project-name={project_name} --namespace {namespace}
```

> **NOTE:** If the Secret is not available during the deprovision action, cleaning up the resources created on the GCP side is skipped.

Use the following environment variables to configure the script:

| Name | Required | Default | Description |
|-----|---------|--------|------------|
| **WORKING_NAMESPACE** | YES | - | Name of the Namespace where the GCP Service Broker is installed. |


Use the following flags to configure the application:

| Name | Required |  Type | Description |
|-----|-----|--------|------------|
| **action** | YES | enum | The name of the action to perform. The possible values are: `provision` or `deprovision`. |

Example command:
```bash
env WORKING_NAMESPACE=stage ./bin/gcp-broker.sh --action provision
```

The example output looks as follows:
```bash
kubectl version
Client Version: v1.10.0
Server Version: v1.10.7

gcloud version
Google Cloud SDK 225.0.0
bq 2.0.37
core 2018.11.09
gsutil 4.34

sc version
sc version 0.1.1 linux/amd64

Activated service account credentials for: [adamw-test-account@kyma-project.iam.gserviceaccount.com]
using project:  kyma-project
enabling a GCP API: servicebroker.googleapis.com
enabling a GCP API: bigquery-json.googleapis.com
enabling a GCP API: bigtableadmin.googleapis.com
enabling a GCP API: ml.googleapis.com
enabling a GCP API: pubsub.googleapis.com
enabling a GCP API: spanner.googleapis.com
enabling a GCP API: sqladmin.googleapis.com
enabling a GCP API: storage-api.googleapis.com
enabled required APIs:
  servicebroker.googleapis.com
  bigquery-json.googleapis.com
  bigtableadmin.googleapis.com
  ml.googleapis.com
  pubsub.googleapis.com
  spanner.googleapis.com
  sqladmin.googleapis.com
  storage-api.googleapis.com
generated the key at:  /tmp/service-catalog-gcp983136652/key.json
Broker "default", already exists
Creating k8s resource using 'namespace' template
Creating k8s resource using 'google-oauth-deployment' template
Creating k8s resource using 'service-account-secret' template
Creating k8s resource using 'google-oauth-rbac' template
Creating k8s resource using 'google-oauth-service-account' template
Creating k8s resource using 'gcp-broker' template
The Service Broker has been added successfully.
````

### Run `status-checker.sh` locally

Use the following environment variables to configure the script:

| Name | Required | Default | Description |
|-----|---------|--------|------------|
| **WORKING_NAMESPACE** | YES | - | Name of the Namespace where the GCP Service Broker is installed. |


Use the following flags to configure the application:

| Name  | Required | Type | Description |
|-----|--------|--------|------------|
| **sleep-duration-sec** | NO | integer  | Number of seconds to wait until retrying the status check operation. |
| **max-retries** | NO | integer | The maximum number of retries for checking if the GCP Service Broker was registered in the Service Catalog. |

Example command:
```bash
env WORKING_NAMESPACE=stage ./bin/status-checker.sh --sleep-duration-sec 3 --max-retries 40
```

The example output looks as follows:
```bash
kubectl version
Client Version: v1.10.0
Server Version: v1.10.7

Checking if GCP ServiceBroker is registered in Service Catalog...
Error from server (NotFound): servicebrokers.servicecatalog.k8s.io "gcp-broker" not found
[Retry 1/40] GCP ServiceBroker condition Ready is not set to True - retry in 3s
Error from server (NotFound): servicebrokers.servicecatalog.k8s.io "gcp-broker" not found
[Retry 2/40] GCP ServiceBroker condition Ready is not set to True - retry in 3s
Error from server (NotFound): servicebrokers.servicecatalog.k8s.io "gcp-broker" not found
[Retry 3/40] GCP ServiceBroker condition Ready is not set to True - retry in 3s
[Retry 4/40] GCP ServiceBroker condition Ready is not set to True - retry in 3s
GCP ServiceBroker is registered successfully.
```
 
### Build a Docker image

The GCP Broker Provider image has the following binaries installed:

* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [gcloud](https://cloud.google.com/sdk/)
* [sc](https://github.com/kyma-incubator/k8s-service-catalog)

For more details, see [this](Dockerfile) Dockerfile.

To build a local Docker image, run this command:
```bash
docker build --no-cache -t gcp-broker-provider
```

You can override versions of the kubectl and sc tools using these build arguments:
```bash
--build-arg KUBECTL_CLI_VERSION={KUBECTL_CLI_VERSION}
--build-arg SC_CLI_VERSION={SC_CLI_VERSION}
```
 
