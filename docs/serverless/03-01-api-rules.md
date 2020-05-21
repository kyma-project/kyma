---
title: Exposing Functions
type: Details
---

By default, Functions in Kyma are not exposed outside the cluster. The Knative Serving Controller normally creates two virtual services, one for the internal cluster communication and the other one for exposing the Function outside the cluster. To restrict the access, the Function Controller automatically sets the default `serving.knative.dev/visibility=cluster-local` label on a KService CR that is added to its metadata when a Function CR is processed. This limits the Knative Serving Controller to create only one, local virtual service that allows only for cluster-wide access to Function services. Other resources can access such a Function within the cluster under the `{service-name}.{namespace}.svc.cluster.local` endpoint, such as `test-function.default.svc.cluster.local`.

> **TIP:** For more details on cluster-local services in Knative, read [this](https://knative.dev/docs/serving/cluster-local-route/) document.

To expose a Function outside the cluster, you must create an [APIRule custom resource (CR)](/components/api-gateway#custom-resource-api-rule):

![Expose a Function service](./assets/api-rules.svg)

1. Create the APIRule CR where you specify the Function to expose, define a [Oathkeeper Access Rule](/components/api-gateway/#details-available-security-options) to secure it, and list which HTTP request methods you want to enable for it.

> **CAUTION:** If you decide to expose your Function on an unsecured endpoint, use the `noop` **handler** for the **accessStrategy** you define in the APIRule CR. The `allow` value for the **handler** is not supported in the current Serverless implementation.

2. The API Gateway Controller detects a new APIRule CR and reads its definition.

3. The API Gateway Controller creates an Istio Virtual Service and Access Rules according to details specified in the CR. Such a Function service is available under the `{host-name}.{domain}` endpoint, such as `my-function.kyma.local`.

This way you can specify multiple API Rules with different authentication methods for a single Function service.

> **TIP:** See [this](#tutorials-expose-a-function-with-an-api-rule) tutorial for a detailed example.
