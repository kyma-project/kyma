---
title: Istio sidecars in Kyma and why you want them
---

Istio installed as part of Kyma has [default configuration](./smsh-02-default-istio-setup-in-kyma.md) with automatic Istio proxy sidecar injection enabled by default. It means that all pods of your workloads (Deployments, StatefulSets and other) will get their own sidecar proxy container running next to your application (provided your workload is not in `istio-system` or `kube-system` namespace). This is a requirement for the resource to be a part of Istio service mesh. It has numerous benefits, mentioned below, that you get out of the box, but are complex to manage without this feature.

## Secure communication

One of [Kyma's default configuration option](./smsh-02-default-istio-setup-in-kyma.md) is [Peer authentication](https://istio.io/latest/docs/concepts/security/#peer-authentication) set to `STRICT` mode cluster-wide. This ensures that your workload only accepts [mutual TLS traffic](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/) where both, client and server certificates are validated having all traffic encrypted. This provides each service a strong identity with reliable key and certificate management system.

Another security benefit of having sidecar proxy is possibility to perform request authentication for your service. Istio enables [Request Authentication](https://istio.io/latest/docs/reference/config/security/request_authentication/) using JSON Web Token (JWT) validation using a custom authentication provider. You can learn how to secure your service using JWT in [this tutorial](../../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-workload-jwt.md).

## Observability

Another benefit of having Istio proxies is related to tracing. Istio is performing global tracing and forwarding acquired data to the [Kyma's tracing component](../../../01-overview/main-areas/observability/obsv-03-tracing-in-kyma.md) using [Zipkin protocol](https://zipkin.io). You can read more about the process in the [Tracing Architecture](../../../05-technical-reference/00-architecture/obsv-03-architecture-tracing.md) section.

Kiali is another tool that comes as [separate Kyma component](../../../05-technical-reference/00-architecture/obsv-04-architecture-kiali.md) and Kyma configures Istio to export metrics necessary to support Kiali features that facilitates managing, visualizing and troubleshooting your mesh.

Moreover, Kyma provides [Istio specific dashboards](https://istio.io/latest/docs/ops/integrations/grafana/#configuration) for [Grafana (monitoring component)](../../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring.md) and together with metrics exposed by the Istio sidecar provides better visibility into workloads and mesh control plane performance.

Being part of Istio service mesh enables all these advanced observability features, which would not be possible without advanced instrumentation code within your application.

## Traffic management

[Traffic management](https://istio.io/latest/docs/concepts/traffic-management/) is an important feature of service mesh and having the sidecar injected into every workload makes this feature available without additional configuration.

[Traffic Shifting](https://istio.io/latest/docs/tasks/traffic-management/traffic-shifting/) and [Request Routing](https://istio.io/latest/docs/tasks/traffic-management/request-routing/) allows developers to release their software using techniques like canary releases and A/B testing to make release process faster and more reliable.

[Mirroring](https://istio.io/latest/docs/tasks/traffic-management/mirroring/) and [Fault Injection](https://istio.io/latest/docs/tasks/traffic-management/fault-injection/) can be used for testing and audit purposes to improve resiliency of your applications.

### Resiliency

Application resiliency is an important topic within traffic management functionalities. Traditionally resiliency features like timeouts, retries and circuit breakers were implemented by application libraries, however with service mesh you can delegate that tasks to the mesh and the same configuration options will work regardless of the programming language of your application. You can read more about it in [Network resilience and testing](https://istio.io/latest/docs/concepts/traffic-management/#network-resilience-and-testing)

## Sidecar proxy behavior during Kyma upgrade.

When upgrading Kyma we want to have full compatibility of existing workloads with the upgraded version of Istio. To achieve that we try to perform `rollout restart` of the workloads whenever possible because this ensures that newest version of sidecar proxy will be injected into the pods. There are exceptions when it's not possible to restart workloads and this behaviour is described in [troubleshooting guide](https://kyma-project.io/docs/kyma/latest/04-operation-guides/troubleshooting/apix-09-upgrade-sidecar-proxy#cause) 
