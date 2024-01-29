# Troubleshooting

The troubleshooting section aims to identify the most common recurring problems the users face when they install and start using Kyma, as well as the most suitable solutions to these problems.

If you can't find a solution, don't hesitate to create a [GitHub](https://github.com/kyma-project/kyma/issues) issue or reach out to our [Slack channel](http://slack.kyma-project.io/) to get direct support from the community.

See the full list of Kyma troubleshooting guides:

## General Troubleshooting Guides

- [Component Doesn't Work After Successful Installation](./01-component-installation-fails.md)
- [Local Kyma Deployment Fails with Pending Pods](./01-deplyoment-fails-pending-pods.md)
- [Provisioning k3d Fails on a Linux Machine](./01-k3d-fails-on-linux.md)
- [Error for kubectl Port Forwarding](./01-kube-troubleshoot-kubectl-port-forward.md)
- [Kyma Domain Is Not Resolvable](./01-kyma-domain-unresolvable.md)
- [Kyma Resource Is Misconfigured](./01-resources-misconfigured.md)
- [Cannot Create a Volume Snapshot](./01-volume-backup.md)

## Istio Module

- [Can’t Access a Kyma Endpoint (503 status code)](https://kyma-project.io/#/istio/user/troubleshooting/03-10-503-no-access)
- [Connection Refused Errors](https://kyma-project.io/#/istio/user/troubleshooting/03-20-connection-refused)
- [Issues with Istio Sidecar Injection](https://kyma-project.io/#/istio/user/troubleshooting/03-30-istio-no-sidecar)
- [Incompatible Istio Sidecar Version After Istio Operator’s Upgrade](https://kyma-project.io/#/istio/user/troubleshooting/03-40-incompatible-istio-sidecar-version)
- [Istio Unintentionally Removed](https://kyma-project.io/#/istio/user/troubleshooting/03-50-recovering-from-unintentional-istio-removal)
- [Kyma Endpoint Returns a not found Error (404 Status Code)](https://kyma-project.io/#/istio/user/troubleshooting/03-60-404-on-istio-gateway)

## Serverless Module

- [Failure to Build Functions](https://kyma-project.io/#/serverless-manager/user/troubleshooting-guides/03-10-cannot-build-functions)
- [Failing Function Container](https://kyma-project.io/#/serverless-manager/user/troubleshooting-guides/03-20-failing-function-container)
- [Functions Failing to Build on k3d](https://kyma-project.io/#/serverless-manager/user/troubleshooting-guides/03-40-function-build-failing-k3d)
- [Serverless Periodically Restarting](https://kyma-project.io/#/serverless-manager/user/troubleshooting-guides/03-50-serverless-periodically-restaring)

## Telemetry Module

- [Traces Troubleshooting](https://kyma-project.io/#/telemetry-manager/user/03-traces?id=troubleshooting)
- [Metrics Troubleshooting](https://kyma-project.io/#/telemetry-manager/user/04-metrics?id=troubleshooting)

## Eventing Module

- [Kyma Eventing - Basic Diagnostics](https://kyma-project.io/#/eventing-manager/user/troubleshooting/evnt-01-eventing-troubleshooting)
- [NATS JetStream Backend Troubleshooting](https://kyma-project.io/#/eventing-manager/user/troubleshooting/evnt-02-jetstream-troubleshooting)
- [Subscriber Receives Irrelevant Events](https://kyma-project.io/#/eventing-manager/user/troubleshooting/evnt-03-type-collision)
- [Eventing Backend Stopped Receiving Events Due To Full Storage](https://kyma-project.io/#/eventing-manager/user/troubleshooting/evnt-04-free-jetstream-storage)
- [Published Events Are Pending in the Stream](https://kyma-project.io/#/eventing-manager/user/troubleshooting/evnt-05-fix-pending-messages)

## API Gateway Module

- [Certificate Management - Issuer Not Created](https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-20-cert-mgt-issuer-not-created)
- [Kyma Gateway - Not Reachable](https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-30-gateway-not-reachable)
- [Issues When Creating an APIRule - Various Reasons](https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-40-api-rule-troubleshooting)
- [Issues with Certificates on Gardener](https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-50-certificates-gardener)
- [Cannot Connect to a Service Exposed by an APIRule - Basic Diagnostics](https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-00-cannot-connect-to-service/03-00-basic-diagnostics)
- [DNS Management](https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-10-dns-mgt/README)
  - [Connection Refused or Timeout](https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-10-dns-mgt/03-10-dns-mgt-connection-refused)
  - [Could Not Resolve Host](https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-10-dns-mgt/03-11-dns-mgt-could-not-resolve-host)
  - [Resource Ignored by the Controller](https://kyma-project.io/#/api-gateway/user/troubleshooting-guides/03-10-dns-mgt/03-12-dns-mgt-resource-ignored)

## Security

- [Issues with Certificates on Gardener](./security/sec-01-certificates-gardener.md)
