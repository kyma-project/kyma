# Helm Broker

## Overview

The Helm Broker is an implementation of a service broker which runs on the Kyma cluster and deploys applications into the Kubernetes cluster using Kyma bundles, and the [Helm](https://github.com/kubernetes/helm) client. A bundle is an abstraction layer over a Helm chart which allows you to represent it as a ClusterServiceClass in the Service Catalog. For example, a bundle can provide plan definitions or binding details. The Helm Broker fetches bundle definitions from an HTTP server. By default, the Helm Broker contains an embedded HTTP server which serves bundles from the Kyma bundles directory.

For the details about Helm Broker configuration, see [this](https://github.com/kyma-project/kyma/blob/master/docs/service-brokers/docs/011-configuration-helm-broker.md) document. See [How to create a yBundle](https://github.com/kyma-project/kyma/blob/master/docs/service-brokers/docs/012-configuration-helm-broker-bundles.md) and [Binding yBundles](https://github.com/kyma-project/kyma/blob/master/docs/service-brokers/docs/013-configuration-helm-broker-bundles-binding.md) to learn more about the bundles.
The Helm Broker implements the Service Broker API. For more information about the Service Brokers, see the [Service Brokers](https://github.com/kyma-project/kyma/blob/master/docs/service-brokers/docs/001-overview-service-brokers.md) overview document.

## Prerequisites

You need the following tools to set up the project:
* The 1.9 or higher version of [Go](https://golang.org/dl/)
* The latest version of [Docker](https://www.docker.com/)
* The latest version of [Dep](https://github.com/golang/dep)


## Development

Before each commit, use the `before-commit.sh` script, which tests your changes.
