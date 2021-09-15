This set of Get Started guides will show you how to set sail with Kyma and demonstrate its main use cases.

All guides, whenever possible, demonstrate the steps in both kubectl and Kyma Dashboard.
All the steps are performed in the `deafult` Namespace.

## Prerequisites

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) (1.19 or greater)
- [curl](https://github.com/curl/curl)
- [k3d](https://k3d.io/#installation)
- [Kyma CLI](../04-operation-guides/operations/01-install-kyma-CLI.md)

## Steps

These guides cover the following steps:

1. [Quick install](01-quick-install.md), which shows how to quickly provision a Kyma cluster locally using k3d.
2. [Deploy and expose a Function](02-deploy-expose-function.md), which shows how to deploy a sample function in a matter of seconds and how to expose it through the APIRule custom resource (CR) on HTTP endpoints. This way it will be available for other services outside the cluster.
3. [Deploy and expose a microservice](03-deploy-expose-microservice.md), which demonstrates how to create a sample microservice and, as before, how to expose it so that it is available for other services outside the cluster.
4. [Trigger your workload with an event](04-trigger-workload-with-event.md), which shows how to trigger your Function or microservice with a sample event.
5. [Observability](05-observability.md), which shows how to access the Grafana dashboard for the Function and fetch the logs.

Let's get started!