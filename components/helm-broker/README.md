# Helm Broker

## Overview

The Helm Broker is an implementation of a Service Broker which runs on the Kyma cluster and deploys applications into the Kubernetes cluster using Kyma bundles, and the [Helm](https://github.com/kubernetes/helm) client. A bundle is an abstraction layer over a Helm chart which allows you to represent it as a ClusterServiceClass in the Service Catalog. For example, a bundle can provide plan definitions or binding details. The Helm Broker fetches bundles definitions from an HTTP servers. A list of HTTP bundles repositories is defined in the ConfigMap and can be changed in the runtime.

For the details about Helm Broker configuration, see [this](../../docs/helm-broker/08-01-configure-hb.md) document. See the [Create a bundle](../../docs/helm-broker/03-01-create-bundles.md) and [Bind bundles](../../docs/helm-broker/03-02-bind-bundles.md) documents to learn more about bundles.
The Helm Broker implements the Service Broker API. For more information about the Service Brokers, see the [Service Brokers](../../docs/service-catalog/13-01-service-brokers.md) overview document.

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
APP_KUBECONFIG_PATH=/Users/$User/.kube/config APP_CONFIG_FILE_NAME=contrib/minimal-config.yaml  APP_CLUSTER_SERVICE_BROKER_NAME=helm-broker APP_HELM_BROKER_URL=http://localhost:8080 APP_NAMESPACE=kyma-system go run cmd/broker/main.go
```

>**NOTE:**  Not all features are available when you run the Helm Broker locally. All features which perform actions with Tiller do not work.
