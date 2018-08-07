<p align="center">
<img src="https://raw.githubusercontent.com/kyma-project/kyma/master/logo.png" width="235">
</p>

## Overview

Kyma is a cloud-native application development framework.

It provides the last mile capabilities that a developer needs to build a cloud-native application using several open-source projects under the Cloud Native Computing Foundation (CNCF), such as Kubernetes, Istio, NATS, Kubeless, and Prometheus, to name a few.
It is designed natively on Kubernetes and, therefore, it is portable to all major cloud providers.

Kyma allows you to connect and extend products in a quick and modern way, using serverless computing and microservice architecture.

The extensions and customizations you create are decoupled from the core applications, which means that:
* deployments are quick
* scaling is independent from the core applications
* the changes you make can be easily reverted without causing downtime of the production system

Living outside of the core product, Kyma allows you to be completely language-agnostic and customize your solution using the technology stack you want to use, not the one the core product dictates. Additionally, Kyma follows the "batteries included" principle and comes with all of the "plumbing code" ready to use, allowing you to focus entirely on writing the domain code and business logic.

Read the [documentation](docs/README.md) to learn about the product, its technology stack, and components.

## Installation

Install Kyma [locally](docs/kyma/docs/031-gs-local-installation.md) and on a [cluster](docs/kyma/docs/032-gs-cluster-installation.md).

## Usage

Kyma comes with the ready-to-use code snippets that you can use to test the extensions and the core functionality. See the list of existing examples in the [`examples`](https://github.com/kyma-project/examples) repository.
