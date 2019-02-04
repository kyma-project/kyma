# Helm Broker

## Overview

The Helm Broker is an implementation of a service broker which runs on the Kyma cluster and deploys applications into the Kubernetes cluster using Kyma bundles, and the [Helm](https://github.com/kubernetes/helm) client. A bundle is an abstraction layer over a Helm chart which allows you to represent it as a ClusterServiceClass in the Service Catalog. For example, a bundle can provide plan definitions or binding details. The Helm Broker fetches bundle definitions from an HTTP server. By default, the Helm Broker contains an embedded HTTP server which serves bundles from the Kyma bundles directory.

For the details about Helm Broker configuration, see [this](../../docs/service-brokers/docs/05-01-helm-broker.md) document. See [How to create a bundle](../../docs/service-brokers/docs/05-02-helm-broker-bundles.md) and [Binding bundles](../../docs/service-brokers/docs/05-03-helm-broker-bundles-binding.md) to learn more about the bundles.
The Helm Broker implements the Service Broker API. For more information about the Service Brokers, see the [Service Brokers](../../docs/service-brokers/docs/01-01-service-brokers.md) overview document.

## Prerequisites

You need the following tools to set up the project:
* The 1.9 or higher version of [Go](https://golang.org/dl/)
* The latest version of [Docker](https://www.docker.com/)
* The latest version of [Dep](https://github.com/golang/dep)


## Development

Before each commit, use the `before-commit.sh` script, which tests your changes.

### Run a local version

To run the application without building a binary file, run this command:

```bash
APP_KUBECONFIG_PATH=/Users/$User/.kube/config APP_CONFIG_FILE_NAME=contrib/minimal-config.yaml APP_REPOSITORY_URLS=https://github.com/kyma-project/bundles/releases/download/0.1.0/ APP_CLUSTER_SERVICE_BROKER_NAME=helm-broker APP_HELM_BROKER_URL=http://localhost:8080 go run cmd/broker/main.go
```

>**NOTE:**  Not all features are available when you run the Helm Broker locally. All features which perform actions with Tiller do not work.
