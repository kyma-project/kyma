# Helm Broker

## Overview

The Helm Broker is a [Service Broker](https://kyma-project.io/docs/master/components/service-catalog/#service-brokers-overview) which exposes Helm charts as Service Classes in the [Service Catalog](https://kyma-project.io/docs/master/components/service-catalog/#overview-overview). To do so, the Helm Broker uses the concept of addons. An addon is an abstraction layer over a Helm chart which provides all information required to convert the chart into a Service Class.

For more information, read the [Helm Broker documentation](https://kyma-project.io/docs/master/components/helm-broker/).

## Prerequisites

To set up the project, download these tools:

* [Go](https://golang.org/dl/) 1.11.4
* [Dep](https://github.com/golang/dep) v0.5.0
* [Docker](https://www.docker.com/)

These Go and Dep versions are compliant with the `buildpack` used by Prow. For more details read [this](https://github.com/kyma-project/test-infra/blob/master/prow/images/buildpack-golang/README.md) document.

## Development

Before each commit, use the `before-commit.sh` script, which tests your changes.

### Run a local version

To run the application without building a binary file, run this command:

```bash
APP_KUBECONFIG_PATH=/Users/$User/.kube/config APP_CONFIG_FILE_NAME=contrib/minimal-config.yaml  APP_CLUSTER_SERVICE_BROKER_NAME=helm-broker APP_HELM_BROKER_URL=http://localhost:8080 APP_NAMESPACE=kyma-system go run cmd/broker/main.go
```

>**NOTE:**  Not all features are available when you run the Helm Broker locally. All features which perform actions with Tiller do not work.
