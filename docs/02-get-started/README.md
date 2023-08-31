---
title: Get Started
---

This set of Get Started guides shows you how to set sail with Kyma and demonstrates its main use cases.

All guides, whenever possible, demonstrate the steps in both kubectl and Kyma Dashboard.
All the steps are performed in the `default` Namespace.

## Prerequisites <!-- {docsify-ignore} -->

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) (v1.26 or higher)
- [curl](https://github.com/curl/curl)
- [k3d](https://k3d.io) (v5.4.9 or higher with [a Kubernetes version supported by Kyma](../04-operation-guides/operations/02-install-kyma.md))
- [Kyma CLI](../04-operation-guides/operations/01-install-kyma-CLI.md)
- Minimum Docker resources: 4 CPUs and 8 GB RAM 
  > **NOTE:** Learn how to adjust these Docker values on [Mac](https://docs.docker.com/desktop/settings/mac/#resources), [Windows](https://docs.docker.com/desktop/settings/windows/#resources), or [Linux](https://docs.docker.com/desktop/settings/linux/#resources).
- (Optional) [CloudEvents Conformance Tool](https://github.com/cloudevents/conformance) for [triggering workloads with events](./04-trigger-workload-with-event.md)
   ```bash
   go install github.com/cloudevents/conformance/cmd/cloudevents@latest
   ```
  
    Alternatively, you can just use `curl` to publish events.
- Istio sidecar injection enabled in the `default` Namespace
  >**NOTE:** Read about [Istio sidecars in Kyma and why you want them](../01-overview/service-mesh/smsh-03-istio-sidecars-in-kyma.md). Then, check how to [enable automatic Istio sidecar proxy injection](../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md). For more details, see [Default Istio setup in Kyma](../01-overview//service-mesh/smsh-02-default-istio-setup-in-kyma.md).

## Steps <!-- {docsify-ignore} -->

These guides cover the following steps:

1. [Deploy and expose a Function](02-deploy-expose-function.md), which shows how to deploy a sample function in a matter of seconds and how to expose it through the APIRule custom resource (CR) on HTTP endpoints. This way it will be available for other services outside the cluster.
2. [Deploy and expose a microservice](03-deploy-expose-microservice.md), which demonstrates how to create a sample microservice and, as before, how to expose it so that it is available for other services outside the cluster.
3. [Trigger your workload with an event](04-trigger-workload-with-event.md), which shows how to trigger your Function or microservice with a sample event.
4. [Observability](05-observability.md), which shows how to access the Grafana dashboard and view the logs and metrics for the Function.

Let's get started!
