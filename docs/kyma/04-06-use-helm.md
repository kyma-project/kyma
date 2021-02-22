---
title: Use Helm
type: Installation
---

You can use Helm to manage Kubernetes resources in Kyma, for example to check the already installed Kyma charts or to install new charts.

## Helm v3

As of version 1.14, Kyma uses [Helm v3](https://helm.sh/) to deploy and maintain components. Unlike its predecessor, Helm v3 interacts directly with the Kubernetes API and thus no longer features an in-cluster server. With Tiller gone, managing Kubernetes resources using Helm v3 CLI requires no manual configuration.
