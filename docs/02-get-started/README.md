---
title: Get Started
---

This set of Get Started guides will show you how to set sail with Kyma and demonstrate its main use cases.

All guides, whenever possible, demonstrate the steps in both kubectl and Kyma Dashboard.
All the steps are performed in the `default` Namespace.

## Prerequisites

>**CAUTION:** As of version 1.20, [Kubernetes deprecated Docker](https://kubernetes.io/blog/2020/12/02/dont-panic-kubernetes-and-docker/) as a container runtime in favor of [containerd](https://containerd.io/). Due to a different way in which [containerd handles certificate authorities](https://github.com/containerd/containerd/issues/3071), Kyma's built-in Docker registry does not work correctly on clusters running with a self-signed TLS certificate on top of Kubernetes installation where containerd is used as a container runtime. If that is your case, either upgrade the cluster to use Docker instead of containerd, generate a valid TLS certificate for your Kyma instance or [configure an external Docker registry](https://kyma-project.io/docs/kyma/latest/03-tutorials/00-serverless/svls-07-set-external-registry/).

- [Kubernetes](https://kubernetes.io/docs/setup/) (v1.19 or greater)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) (v1.19 or greater)
- [curl](https://github.com/curl/curl)
- [k3d](https://k3d.io) (v4.0.0 or greater)
- [Kyma CLI](../04-operation-guides/operations/01-install-kyma-CLI.md)

## Steps

These guides cover the following steps:

1. [Quick install](01-quick-install.md), which shows how to quickly provision a Kyma cluster locally using k3d.
2. [Deploy and expose a Function](02-deploy-expose-function.md), which shows how to deploy a sample function in a matter of seconds and how to expose it through the APIRule custom resource (CR) on HTTP endpoints. This way it will be available for other services outside the cluster.
3. [Deploy and expose a microservice](03-deploy-expose-microservice.md), which demonstrates how to create a sample microservice and, as before, how to expose it so that it is available for other services outside the cluster.
4. [Trigger your workload with an event](04-trigger-workload-with-event.md), which shows how to trigger your Function or microservice with a sample event.
5. [Observability](05-observability.md), which shows how to access the Grafana dashboard and view the logs and metrics for the Function.

Let's get started!
