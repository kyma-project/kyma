---
title: Troubleshooting
---

The troubleshooting section aims to identify the most common recurring problems the users face when they install and start using Kyma, as well as the most suitable solutions to these problems.

If you can't find a solution, don't hesitate to create a [GitHub](https://github.com/kyma-project/kyma/issues) issue or reach out to our [Slack channel](http://slack.kyma-project.io/) to get direct support from the community.

See the full list of Kyma troubleshooting guides:

### General troubleshooting guides

- [Component doesn't work after successful installation](./01-component-installation-fails.md)
- [Local Kyma deployment fails with pending Pods](./01-deplyoment-fails-pending-pods.md)
- [Provisioning k3d fails on a Linux machine](./01-k3d-fails-on-linux.md)
- [Error for kubectl port forwarding](./01-kube-troubleshoot-kubectl-port-forward.md)
- [Kyma domain is not resolvable](./01-kyma-domain-unresolvable.md)
- [Kyma resource is misconfigured](./01-resources-misconfigured.md)
- [Cannot create a volume snapshot](./01-volume-backup.md)

### API Exposure
  
- Cannot connect to a service exposed by an APIRule
  - [Basic diagnostics](./api-exposure/apix-01-cannot-connect-to-service/apix-01-01-apigateway-connect-api-rule.md)
  - [401 Unauthorized or 403 Forbidden](./api-exposure/apix-01-cannot-connect-to-service/apix-01-02-401-unauthorized-403-forbidden.md)
  - [404 Not Found](./api-exposure/apix-01-cannot-connect-to-service/apix-01-03-404-not-found.md)
  - [500 Internal Server Error](./api-exposure/apix-01-cannot-connect-to-service/apix-01-04-500-server-error.md)
- External DNS management
  - [Connection refused or timeout](./api-exposure/apix-02-dns-mgt/apix-02-01-dns-mgt-connection-refused.md)
  - [Could not resolve host](./api-exposure/apix-02-dns-mgt/apix-02-02-dns-mgt-could-not-resolve-host.md)
  - [Resource ignored by the controller](./api-exposure/apix-02-dns-mgt/apix-02-03-dns-mgt-resource-ignored.md)
- [Certificate management - Issuer not created](./api-exposure/apix-03-cert-mgt-issuer-not-created.md)
- [Kyma Gateway - not reachable](./api-exposure/apix-04-gateway-not-reachable.md)
- [Pods stuck in `Pending/Failed/Unknown` state after an upgrade](./api-exposure/apix-05-upgrade-sidecar-proxy.md)
- [Issues when creating an APIRule - various reasons](./api-exposure/apix-06-api-rule-troubleshooting.md)

### Eventing

- [Kyma Eventing - Basic Diagnostics](./eventing/evnt-01-eventing-troubleshooting.md)
- [NATS JetStream backend troubleshooting](./eventing/evnt-02-jetstream-troubleshooting.md)
- [Subscriber receives irrelevant events](./eventing/evnt-03-type-collision.md)
- [Eventing backend stopped receiving events due to full storage](./eventing/evnt-04-free-jetstream-storage.md)
- [Published events are pending in the stream](./eventing/evnt-05-fix-pending-messages.md)

### Observability

- [Prometheus Istio Server restarting or in crashback loop](./observability/obsv-01-troubleshoot-prometheus-istio-server-crash-oom.md)
- [Component doesn't work after successful installation](./observability/obsv-02-troubleshoot-trace-backend-shows-few-traces.md)

### Security
  
- [Issues with certificates on Gardener](./security/sec-01-certificates-gardener.md)

### Serverless

- [Failure to build Functions](./serverless/svls-01-cannot-build-functions.md)
- [Failing Function container](./serverless/svls-02-failing-function-container.md)
- [Function debugger stops at dependency files](./serverless/svls-03-function-debugger-in-strange-location.md)
- [Functions failing to build on k3d](./serverless/svls-04-function-build-failing-on-k3d.md)
- [Serverless periodically restarting](./serverless/svls-05-serverless-periodically-restarting.md)

### Service Mesh

- [Can't access a Kyma endpoint (503 status code)](./service-mesh/smsh-01-503-no-access.md)
- [Connection refused errors](./service-mesh/smsh-02-connection-refused.md)
- [Issues with Istio sidecar injection](./service-mesh/smsh-03-istio-no-sidecar.md)
- [Incompatible Istio sidecar version after Kyma upgrade](./service-mesh/smsh-04-istio-sidecar-version.md)
