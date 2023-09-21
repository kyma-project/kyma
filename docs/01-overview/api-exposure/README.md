---
title: What is API Exposure in Kyma?
---

The API exposure in Kyma is based on the API Gateway component that aims to provide a set of functionalities that allow developers to expose, secure, and manage their APIs in an easy way. The main element of the API Gateway is the API Gateway Controller, which exposes services in Kyma.

# API Gateway

To make your service accessible outside the Kyma cluster, expose it using Kyma API Gateway Controller, which listens for the custom resource (CR) objects that follow the `apirules.gateway.kyma-project.io` CustomResourceDefinition (CRD). Creating a valid CR triggers API Gateway Controller to create an Istio VirtualService. Optionally, you can specify the **rules** attribute of the CR to secure the exposed service with Oathkeeper Access Rules.

API Gateway Controller allows you to secure the exposed services using JWT tokens issued by an OpenID Connect-compliant identity provider, or OAuth2 tokens issued by the Kyma OAuth2 server. You can secure the entire service, or secure the selected endpoints. Alternatively, you can leave the service unsecured.

>**CAUTION:** Since Kyma 2.2, Ory stack has been deprecated, and Ory Hydra was removed with Kyma 2.19. For more information, read the blog posts explaining the [new architecture](https://blogs.sap.com/2023/02/10/sap-btp-kyma-runtime-api-gateway-future-architecture-based-on-istio/) and [Ory Hydra migration](https://blogs.sap.com/2023/06/06/sap-btp-kyma-runtime-ory-hydra-oauth2-client-migration/). See the [deprecation note](https://github.com/kyma-project/website/blob/main/content/blog-posts/2022-05-04-release-notes-2.2/index.md#ory-stack-deprecation-note).


# API Gateway limitations

## Controller limitations

The APIRule controller is not a critical component of the Kyma networking infrastructure since it relies on Istio and Ory Custom Resources to provide routing capabilities. In terms of persistence, the controller depends on APIRules stored in the Kubernetes cluster. No other persistence solution is present.

In terms of the resource configuration, the following requirements are set on the API Gateway controller:

|          | CPU  | Memory |
|----------|------|--------|
| Limits   | 100m | 128Mi  |
| Requests | 10m  | 64Mi   |

## Limitations in terms of the number of created APIRules

The number of created APIRules is not limited. 

## Dependencies

API Gateway depends on Istio and Ory to provide its routing capabilities. In the case of the `allow` access strategy, only a Virtual Service Custom Resource is created. With any other access strategy, both Virtual Service and Oathkeeper Rule Custom Resource are created.

# Useful links

If you're interested in learning more about API Exposure in Kyma, follow these links to:

- Perform some simple and more advanced tasks:
  - [Create a workload](../../03-tutorials/00-api-exposure/apix-01-create-workload.md)
  - [Set up a custom domain for a workload](../../03-tutorials/00-api-exposure/apix-02-setup-custom-domain-for-workload.md)
  - [Expose a workload](../../03-tutorials/00-api-exposure/apix-04-expose-workload/apix-04-01-expose-workload-apigateway.md)
  - [Expose multiple workloads on the same host](../../03-tutorials/00-api-exposure/apix-04-expose-workload/apix-04-02-expose-multiple-workloads.md)
  - [Expose and secure a workload with OAuth2](../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2.md)
  - [Get a JWT](../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-02-get-jwt.md)
  - [Expose and secure a workload with Istio](../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-04-expose-and-secure-workload-istio.md)
  - [Expose and secure a workload with JWT](../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt.md)
  - [Activate access logs](../../04-operation-guides/operations/obsv-03-enable-istio-access-logs.md)

- Troubleshoot API Exposure-related issues when:

  - You [cannot connect to a service exposed by an APIRule](../../04-operation-guides/troubleshooting/api-exposure/apix-01-cannot-connect-to-service/apix-01-01-apigateway-connect-api-rule.md)
  - You get the [`404 Not Found`](../../04-operation-guides/troubleshooting/api-exposure/apix-01-cannot-connect-to-service/apix-01-03-404-not-found.md) status code when you try to connect to a service exposed by an APIRule
  - You get the [`500 Internal Server Error`](../../04-operation-guides/troubleshooting/api-exposure/apix-01-cannot-connect-to-service/apix-01-04-500-server-error.md) status code when you try to connect to a service exposed by an APIRule
  - [Connection refused](../../04-operation-guides/troubleshooting/api-exposure/apix-02-dns-mgt/apix-02-01-dns-mgt-connection-refused.md) errors occur when you want to use your custom domain
  - You receive the [`could not resolve host`](../../04-operation-guides/troubleshooting/api-exposure/apix-02-dns-mgt/apix-02-02-dns-mgt-could-not-resolve-host.md) error when you want to use your custom domain
  - A [resource is ignored by the controller](../../04-operation-guides/troubleshooting/api-exposure/apix-02-dns-mgt/apix-02-03-dns-mgt-resource-ignored.md)
  - The [Issuer Custom Resource fails to be created](../../04-operation-guides/troubleshooting/api-exposure/apix-03-cert-mgt-issuer-not-created.md)
  - The [Kyma Gateway is not reachable](../../04-operation-guides/troubleshooting/api-exposure/apix-04-gateway-not-reachable.md)
  - The [Pods are stuck in `Pending/Failed/Unknown` state after an upgrade](../../04-operation-guides/troubleshooting/api-exposure/apix-05-upgrade-sidecar-proxy.md)
  - There are [Issues when creating an APIRule - Various reasons](../../04-operation-guides/troubleshooting/api-exposure/apix-06-api-rule-troubleshooting.md)

- Learn something more about:

  - [Authorization configuration](../../05-technical-reference/apix-01-config-authorizations-apigateway.md)
  - [Allowed domains in API Gateway Controller](../../05-technical-reference/apix-02-whitelisted-domains.md)
  - [Blocked services in API Gateway Controller](../../05-technical-reference/apix-03-blacklisted-services.md)

- Analyze configuration details for:

  - The [ORY chart](../../05-technical-reference/00-configuration-parameters/apix-02-ory-chart.md)
