---
title: What is Service Mesh is Kyma?
---

Kyma Service Mesh is an infrastructure layer that handles service-to-service communication, proxying, service discovery, traceability, and security, independently of the code of the services. To deliver this functionality, Kyma uses [Istio](https://istio.io/docs/concepts/what-is-istio/) Service Mesh that is customized for the specific needs of the implementation.

## Details

The main principle of Kyma Service Mesh is to inject Pods of every service with the Envoy sidecar proxy. Envoy intercepts the communication between the services and regulates it by applying and enforcing the rules you create.

By default, Istio in Kyma has [mutual TLS (mTLS)](https://istio.io/docs/concepts/security/#mutual-tls-authentication) disabled. See how to [enable sidecar proxy injection](../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md). You can manage mTLS traffic in services or at a Namespace level by creating [DestinationRules](https://istio.io/docs/reference/config/networking/destination-rule/) and [Peer Authentications](https://istio.io/docs/tasks/security/authentication/authn-policy/). If you disable sidecar injection in a service or in a Namespace, you must manage their traffic configuration by creating appropriate DestinationRules and Peer Authentications.

> **NOTE:** The Istio Control Plane doesn't have mTLS enabled.

> **NOTE:** For security and performance we use the [distroless](https://istio.io/docs/ops/configuration/security/harden-docker-images/) version of Istio images. Those images are not Debian-based and are slimmed down to reduce any potential attack surface and increase startup time.

You can install Service Mesh as part of Kyma predefined [profiles](../../04-operation-guides/operations/02-install-kyma.md#choose-resource-consumption). For production purposes, use the **production profile** which has increased resource quotas for all Istio components. It also has HorizontalPodAutoscaler (HPA) enabled for all Istio components.

# Default Istio setup in Kyma

Istio in Kyma is installed with the help of the `istioctl` tool. The tool is driven by a configuration file containing an instance of the [IstioOperator](https://istio.io/docs/reference/config/istio.operator.v1alpha1/) custom resource.

## Istio components

This list shows the available Istio components and addons. Check which of those are enabled in Kyma:
- Istiod (Pilot)
- Ingress Gateway
- Grafana - installed as separate component - [monitoring](../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring.md)
- Prometheus - installed as separate component - [monitoring](../../05-technical-reference/00-architecture/obsv-01-architecture-monitoring.md)

## Kyma-specific configuration

These configuration changes are applied to customize Istio for use with Kyma:

- Both [Istio control plane and data plane](https://istio.io/latest/docs/ops/deployment/architecture/) use distroless images. To learn more, read about [Harden Docker Container Images](https://istio.io/latest/docs/ops/configuration/security/harden-docker-images/).
- Automatic sidecar injection is disabled by default. See how to [enable sidecar proxy injection](../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md).
- Resource requests and limits for Istio sidecars are modified to best suit the needs of the evaluation and production profiles.
- [Mutual TLS (mTLS)](https://istio.io/docs/concepts/security/#mutual-tls-authentication) is enabled cluster-wide in a STRICT mode.
- Ingress Gateway is expanded to handle ports `80`, `443`, and `31400` for local Kyma deployments.
- The use of HTTP 1.0 is enabled in the outbound HTTP listeners by `PILOT_HTTP10` flag set in Istiod component environment variables.
- IstioOperator configuration file is modified. [Change Kyma settings](../../04-operation-guides/operations/03-change-kyma-config-values.md) to customize the configuration.

# Istio sidecars in Kyma and why you want them

## Purpose of Istio sidecars

By default, Istio installed as part of Kyma [is configured](./smsh-02-default-istio-setup-in-kyma.md) with automatic Istio proxy sidecar injection disabled. This means that none of Pods of your workloads (such as deployments and StatefulSets; except any workloads in the `kyma-system` Namespace) get their own sidecar proxy container running next to your application.

With an Istio sidecar, the resource becomes part of Istio service mesh, which brings the following benefits that would be complex to manage otherwise.



## Secure communication
<!-- markdown-link-check-disable-next-line -->
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

# Istio limitations

## Resource configuration

By default, Istio resources are configured in the following matter:

| Component       |          | CPU   | Memory |
|-----------------|----------|-------|--------|
| Proxy           | Limits   | 1000m | 1024Mi |
| Proxy           | Requests | 10m   | 192Mi  |
| Ingress Gateway | Limits   | 2000m | 1024Mi |
| Ingress Gateway | Requests | 100m  | 128Mi  |
| Pilot           | Limits   | 4000m | 2Gi    |
| Pilot           | Requests | 100m  | 512Mi  |
| CNI             | Limits   | 500m  | 1024Mi |
| CNI             | Requests | 100m  | 512Mi  |

## Autoscaling configuration

The autoscaling configuration of the Istio components is as follows:

| Component       | Min replicas | Max replicas |
|-----------------|--------------|--------------|
| Pilot           | 2            | 5            |
| Ingress Gateway | 3            | 10           |

The CNI component is provided as a DaemonSet, meaning that one replica is present on every node of the target cluster. Istio sidecar proxy isn't configured in terms of autoscaling as it is injected into a Pod with the [sidecar injection enabled](../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md).

# Useful links

If you're interested in learning more about Service Mesh in Kyma, follow these links to:

- Learn about the benefits of having your workload enrolled to the service mesh:
 
  - [Istio sidecars in Kyma and why you want them](./smsh-03-istio-sidecars-in-kyma.md)

- Perform some simple and more advanced tasks:

  - [Enable Istio Sidecar Proxy Injection](../../04-operation-guides/operations/smsh-01-istio-enable-sidecar-injection.md)

- Troubleshoot Service Mesh-related issues when:

  - You [can't access a Kyma endpoint](../../04-operation-guides/troubleshooting/service-mesh/smsh-01-503-no-access.md)
  - [Connection refused errors](../../04-operation-guides/troubleshooting/service-mesh/smsh-02-connection-refused.md) occur
  - [Issues with Istio sidecar injection](../../04-operation-guides/troubleshooting/service-mesh/smsh-03-istio-no-sidecar.md) come up
  - You have an [incompatible Istio sidecar version after Kyma upgrade](../../04-operation-guides/troubleshooting/service-mesh/smsh-04-istio-sidecar-version.md)

- Analyze configuration details for:

   - The [Istio chart](../../05-technical-reference/00-configuration-parameters/smsh-01-istio-chart.md)
   - The [Istio Resources chart](../../05-technical-reference/00-configuration-parameters/smsh-02-istio-resources-chart.md)
