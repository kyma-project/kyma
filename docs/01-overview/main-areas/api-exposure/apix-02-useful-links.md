---
title: Useful links
---

If you're interested in learning more about API Exposure in Kyma, follow these links to:

- Perform some simple and more advanced tasks:
  - [Create a workload](../../../03-tutorials/00-api-exposure/apix-01-create-workload.md)
  - [Set up a custom domain for a workload](../../../03-tutorials/00-api-exposure/apix-02-setup-custom-domain-for-workload.md)
  - [Expose a workload](../../../03-tutorials/00-api-exposure/apix-04-expose-workload/apix-04-01-expose-workload-apigateway.md)
  - [Expose multiple workloads on the same host](../../../03-tutorials/00-api-exposure/apix-04-expose-workload/apix-04-02-expose-multiple-workloads.md)
  - [Expose and secure a workload with OAuth2](../../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-01-expose-and-secure-workload-oauth2.md)
  - [Get a JWT](../../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-02-get-jwt.md)
  - [Expose and secure a workload with Istio](../../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-04-expose-and-secure-workload-istio.md)
  - [Expose and secure a workload with JWT](../../../03-tutorials/00-api-exposure/apix-05-expose-and-secure-a-workload/apix-05-03-expose-and-secure-workload-jwt.md)

- Troubleshoot API Exposure-related issues when:

  - You [cannot connect to a service exposed by an APIRule](../../../04-operation-guides/troubleshooting/api-exposure/apix-01-apigateway-connect-api-rule.md)
  - You get the [`401 Unauthorized` or `403 Forbidden`](../../../04-operation-guides/troubleshooting/api-exposure/apix-02-401-unauthorized-403-forbidden.md) status code when you try to connect to a service exposed by an APIRule
  - You get the [`404 Not Found`](../../../04-operation-guides/troubleshooting/api-exposure/apix-03-404-not-found.md) status code when you try to connect to a service exposed by an APIRule
  - [Connection refused](../../../04-operation-guides/troubleshooting/api-exposure/apix-04-dns-mgt-connection-refused.md) errors occur when you want to use your custom domain
  - You receive the [`could not resolve host`](../../../04-operation-guides/troubleshooting/api-exposure/apix-05-dns-mgt-could-not-resolve-host.md) error when you want to use your custom domain
  - A [resource is ignored by the controller](../../../04-operation-guides/troubleshooting/api-exposure/apix-06-dns-mgt-resource-ignored.md)
  - The [Issuer Custom Resource fails to be created](../../../04-operation-guides/troubleshooting/api-exposure/apix-07-cert-mgt-issuer-not-created.md)
  - The [Kyma Gateway is not reachable](../../../04-operation-guides/troubleshooting/api-exposure/apix-08-gateway-not-reachable.md)
  - The [Pods stuck in `Pending/Failed/Unknown` state after an upgrade](../../../04-operation-guides/troubleshooting/api-exposure/apix-09-upgrade-sidecar-proxy.md)

- Learn something more about:

  - [Authorization configuration](../../../05-technical-reference/apix-01-config-authorizations-apigateway.md)
  - [Allowed domains in API Gateway Controller](../../../05-technical-reference/apix-02-whitelisted-domains.md)
  - [Blocked services in API Gateway Controller](../../../05-technical-reference/apix-03-blacklisted-services.md)

- Analyze configuration details for:

  - The [ORY chart](../../../05-technical-reference/00-configuration-parameters/apix-02-ory-chart.md)
