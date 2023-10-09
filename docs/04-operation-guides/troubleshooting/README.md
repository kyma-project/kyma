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
  - [404 Not Found](./api-exposure/apix-01-cannot-connect-to-service/apix-01-03-404-not-found.md)
  - [500 Internal Server Error](./api-exposure/apix-01-cannot-connect-to-service/apix-01-04-500-server-error.md)
- External DNS management
  - [Connection refused or timeout](./api-exposure/apix-02-dns-mgt/apix-02-01-dns-mgt-connection-refused.md)
  - [Could not resolve host](./api-exposure/apix-02-dns-mgt/apix-02-02-dns-mgt-could-not-resolve-host.md)
  - [Resource ignored by the controller](./api-exposure/apix-02-dns-mgt/apix-02-03-dns-mgt-resource-ignored.md)
- [Certificate management - Issuer not created](./api-exposure/apix-03-cert-mgt-issuer-not-created.md)
- [Kyma Gateway - not reachable](./api-exposure/apix-04-gateway-not-reachable.md)
- [Issues when creating an APIRule - various reasons](./api-exposure/apix-06-api-rule-troubleshooting.md)

### Eventing

- [Kyma Eventing - Basic Diagnostics](./eventing/evnt-01-eventing-troubleshooting.md)
- [NATS JetStream backend troubleshooting](./eventing/evnt-02-jetstream-troubleshooting.md)
- [Subscriber receives irrelevant events](./eventing/evnt-03-type-collision.md)
- [Eventing backend stopped receiving events due to full storage](./eventing/evnt-04-free-jetstream-storage.md)
- [Published events are pending in the stream](./eventing/evnt-05-fix-pending-messages.md)

### Observability

- [Trace backend shows fewer traces than you would like to see](./observability/obsv-02-troubleshoot-trace-backend-shows-few-traces.md)

### Security
  
- [Issues with certificates on Gardener](./security/sec-01-certificates-gardener.md)

### Istio

- [Can't access a Kyma endpoint (503 status code)](https://kyma-project.io/#/istio/user/02-operation-guides/troubleshooting/03-10-503-no-access)
- [Connection refused errors](https://kyma-project.io/#/istio/user/02-operation-guides/troubleshooting/03-20-connection-refused)
- [Issues with Istio sidecar injection](https://kyma-project.io/#/istio/user/02-operation-guides/troubleshooting/03-30-istio-no-sidecar)
- [Incompatible Istio sidecar version after Kyma upgrade](https://kyma-project.io/#/istio/user/02-operation-guides/troubleshooting/03-40-incompatible-istio-sidecar-version)
- [Istio unintentionally deleted](https://kyma-project.io/#/istio/user/02-operation-guides/troubleshooting/03-50-recovering-from-unintentional-istio-module-removal)
