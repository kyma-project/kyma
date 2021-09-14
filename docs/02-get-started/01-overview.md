---
title: Overview
---

This set of Get Started guides will walk you through the major Kyma components and show its main use cases as an Application runtime.



All guides, whenever possible, demonstrate the steps in both kubectl and Kyma Dashboard.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) (1.19 or greater)
- [curl](https://github.com/curl/curl)
- [k3d](https://k3d.io/#installation)
- [Kyma CLI](../04-operation-guides/operations/01-install-kyma-CLI.md)

## Steps

These guides cover the following steps:

1. [Quick install](02-quick-install.md), which shows how to quickly provision a Kyma cluster locally using k3d.
2. [Deploy a Function](03-deploy-function.md), which shows how to deploy a sample function in a matter of seconds.
3. [Expose the Function](04-expose-function.md), which shows how to expose it through the APIRule custom resource (CR) on HTTP endpoints. This way it will be available for other services outside the cluster.
4. [Deploy a microservice](05-deploy-microservice.md), which demonstrates how to create a sample microservice.
5. [Expose the microservice](06-expose-microservice.md), which, as before, shows how to expose it so that it is available for other services outside the cluster.
6. [Trigger your workload with an event](07-trigger-workload-with-event.md), which shows how to trigger your Function or microservice with a sample event.
7. [Observability](08-observability.md), which shows how to fetch the logs, and how to access the Grafana dashboard for your Function.

Let's get started!