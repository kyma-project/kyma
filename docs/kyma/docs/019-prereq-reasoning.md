---
title: Prerequisites reasoning
type: Details
---

## Hyperkit Driver

Minikube can run on different VM drivers. The Hyperkit driver is the only driver on which Kyma is tested, making it a prerequisite.

## Minikube

To work with Kyma, use only the provided installation and deinstallation scripts. Kyma does not work on a basic Minikube cluster that you can start using the `minikube start` command or stop with the `minikube stop` command. If you do not need Kyma on Minikube anymore, remove the cluster with the `minikube delete` command.
