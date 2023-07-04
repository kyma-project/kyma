---
title: Istio sidecars in Kyma and why you want them
---

## Purpose of Istio sidecars

By default, Istio installed as part of Kyma [is configured](./smsh-02-default-istio-setup-in-kyma.md) with automatic Istio proxy sidecar injection disabled. This means that none of Pods of your workloads (such as deployments and StatefulSets; except any workloads in the `kyma-system` Namespace) get their own sidecar proxy container running next to your application.

With an Istio sidecar, the resource becomes part of Istio service mesh, which brings the following benefits that would be complex to manage otherwise.



## Secure communication

In Kyma's [default Istio configuration](./smsh-02-default-istio-setup-in-kyma.md), [peer authentication](https://istio.io/latest/docs/concepts/security/#peer-authentication) is set to cluster-wide `STRICT` mode. This ensures that your workload only accepts [mutual TLS traffic](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/) where both, client and server certificates, are validated to have all traffic encrypted. This provides each service with a strong identity, with reliable key and certificate management system.

Another security benefit of having a sidecar proxy is that you can perform [request authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) for your service. Istio enables request authentication with JSON Web Token (JWT) validation using a custom authentication provider. Learn how to [secure your workload using JWT](../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt.md).

## Observability

Furthermore, Istio proxies improve tracing: Istio performs global tracing and forwards the data to a tracing backend using the [OTLP protocol](https://opentelemetry.io/docs/reference/specification/protocol/). Learn more in [Tracing Architecture](../../01-overview/telemetry/telemetry-03-traces.md#architecture).

Kiali is another tool to monitor the service mesh; and Kyma configures Istio to export metrics necessary to support Kiali features that facilitate managing, visualizing, and troubleshooting your service mesh. Learn more about deploying Kiali to your Kyma cluster in our [Kiali example](https://github.com/kyma-project/examples/tree/main/kiali).

Moreover, Kyma provides [Istio-specific Grafana dashboards](https://istio.io/latest/docs/ops/integrations/grafana/#configuration) for the [monitoring component](../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring.md). Together with metrics exposed by the Istio sidecar, you get better visibility into workloads and the mesh control plane performance.

Being part of Istio service mesh enables all these advanced observability features, which would not be possible without advanced instrumentation code within your application.

## Traffic management

[Traffic management](https://istio.io/latest/docs/concepts/traffic-management/) is an important feature of service mesh. If you have the sidecar injected into every workload, you can use this traffic management without additional configuration.

With [traffic shifting](https://istio.io/latest/docs/tasks/traffic-management/traffic-shifting/) and [request routing](https://istio.io/latest/docs/tasks/traffic-management/request-routing/), developers can use techniques like canary releases and A/B testing to make their software release process faster and more reliable.

To improve the resiliency of your applications, you can use [mirroring](https://istio.io/latest/docs/tasks/traffic-management/mirroring/) and [fault injection](https://istio.io/latest/docs/tasks/traffic-management/fault-injection/) for testing and audit purposes.

### Resiliency

Application resiliency is an important topic within traffic management. Traditionally, resiliency features like timeouts, retries, and circuit breakers were implemented by application libraries. However, with service mesh, you can delegate such tasks to the mesh, and the same configuration options will work regardless of the programming language of your application. You can read more about it in [Network resilience and testing](https://istio.io/latest/docs/concepts/traffic-management/#network-resilience-and-testing).

## Sidecar proxy behavior during Kyma upgrade

For Kyma upgrades, it's a priority to have full compatibility of existing workloads with the upgraded version of Istio. To ensure that the newest version of sidecar proxy is injected into the Pods, the upgrade performs a `rollout restart` of the workloads whenever possible. To learn more about exceptions when it's impossible to restart workloads, read the troubleshooting guide [Pods stuck in `Pending/Failed/Unknown` state after upgrade](https://kyma-project.io/docs/kyma/latest/04-operation-guides/troubleshooting/api-exposure/apix-05-upgrade-sidecar-proxy/#cause).