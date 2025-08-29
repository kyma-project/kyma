# Rate Limit

## Status

Accepted

## Context

It has been decided to introduce rate limit functionality into Kyma to allow users to set rate limit functionality on
the service mesh layer, therefore allowing to consume intended service mesh functionality which is abstracting away
networking concerns outside applications inside the mesh.
Since Istio is an underlying service mesh responsible for the workload networking across Kyma clusters, it is Istio that
allows to configure such a functionality.

In Envoy, there is support for local and global rate limiting. The major difference between local and global rate
limiting is that local rate limiting is applied directly at the sidecar proxy level, controlling the rate of requests on
a per-instance basis.
For consistent rate limiting across your service mesh, the global rate limiting can be utilised.  
The global rate limiting requires a rate limit service that adheres to
the [gRPC RateLimit v3 protocol](https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/ratelimit/v3/rls.proto).
This centralised approach enforces shared stateful rate limits, applicable to all instances and services within the
mesh.

Currently, there are limitations that prevent using Redis on managed Kyma clusters. Because global rate limiting relies
on Redis for persistence, the global rate limiting is not in scope of this ADR.
Additionally, the metrics for rate limiting are also not in scope of this ADR.

## Decision

A new CRD will be introduced in the API Gateway module to allow rate limiting configuration.
In the future, global rate limit might be introduced. The decision has not been made yet, so it has been decided to name
the new CRD `RateLimit` not to indicate that global rate limit will be provided. By naming the current CRD, for
example, `LocalRateLimit,` logical would be to expect also `GlobalRateLimit` to exist at some point.  
The RateLimit CR is used to configure rate limiting on sidecar proxies and Istio Ingress Gateways.
The CRD is structured in a way that allows further extensions, such as introducing global rate limiting.

### Scope of the RateLimit CR

To maintain clarity and simplicity in rate limit configuration, only one RateLimit CR must match a
workload. This ensures a single source of truth for rate limits. Each RateLimit CR includes a default bucket. Allowing
multiple RateLimit CRs would require aggregating multiple default buckets, significantly increasing the complexity of
the underlying EnvoyFilter and making the RateLimit CR behavior more difficult for users to understand. In some cases, it
might not even be possible to aggregate the default buckets.  
Also, additional buckets with `path` criteria in the RateLimit CR must be
merged into one Envoy configuration, which would further complicate the setup.  
During the evaluation, it was observed that configuring multiple Envoy routes with a LocalRateLimit filter led to
unexpected behavior. Specifically, rate limit headers were added multiple times, and the HTTP status code `503` was
returned instead of the expected `429` when a request was rate limited.

The RateLimit CR allows to configure rate limiting on sidecar proxies and Istio Ingress Gateways.  
The workloads that should be rate limited are selected by the required **selectorLabels** field.
The **selectorLabels** field contains a map of labels that indicate a specific set of Pods on which the
configuration should be applied.  
The selector labels are restricted to the namespace in which the RateLimit CR is present. Due to restrictions by Istio
module, the Istio Ingress Gateway can only be deployed to the `istio-system` namespace and the user must create a
RateLimit CR in the `istio-system` namespace to apply rate limits to the Istio Ingress Gateway.

To create a valid EnvoyFilter from the RateLimit CR,
the [PatchContext](https://istio.io/latest/docs/reference/config/networking/envoy-filter/#EnvoyFilter-PatchContext) must
be configured appropriately. The PatchContext must be set to `SIDECAR_INBOUND` for sidecar proxies to ensure that only
ingress traffic is rate limited.
For Istio Ingress Gateway rate limiting, the PatchContext must be set to `GATEWAY`.

### Local Rate Limiting Descriptors Support

The RateLimit CR must support configuration of rate limiting based on the request descriptors. It must be possible to
configure different rate limits for different request descriptors.  
It is also possible to combine request descriptors, for example rate limit on path and headers, in the same rate limit
configuration. In this case the rate limit is applied only if all specified descriptors are present in the request.  
To
support [RateLimit Action](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-ratelimit-action-request-headers),
which is required for the descriptors, the Envoy rate limit filter must be applied to the `HTTP_ROUTE`. This requires
applying a global rate limit filter with a minimalistic configuration to `HTTP_FILTER`.

The RateLimit CR must have a default rate limit bucket configured, as this is required by
the [Envoy Local Rate limit filter](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/local_ratelimit/v3/local_rate_limit.proto#extensions-filters-http-local-ratelimit-v3-localratelimit)
if it's configured on the route level.
This default bucket is used as a fallback for requests that don't match any other local rate limit configuration in the
CR. To support the fallback behavior when multiple local configurations exist, the created EnvoyFilter must
include `always_consume_default_token_bucket: false` in the Local Rate limit filter.

#### Rate Limit by Path

The RateLimit CR allows configuring rate limits for specific paths exposed by the workload. The **local.rateLimits** field
is a list of additional rate limit configurations that contain matching criteria such as the path.

Since a shared bucket between multiple paths is not possible for local rate limiting, each entry in
the **local.rateLimits** list supports only a single path. To avoid confusion, if multiple paths have a similar bucket
configuration, they must be added as separate entries in the **local.rateLimits** list.

There is a limitation with path matching in local rate limiting. Descriptor values are static and do not support
wildcards (`*`), so paths with path or query parameters will be treated as separate paths.
For example, `/path`, `/path*`, and `/path?param=value` will be treated as distinct paths.

Example for EnvoyFilter creation based on the RateLimit CR configuration:

RateLimit CR

```yaml 
apiVersion: ratelimit.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  name: httpbin-local-rate-limit
  namespace: default
spec:
  selectorLabels:
    app: httpbin
  local:
    defaultBucket:
      maxTokens: 100
      tokensPerFill: 50
      fillInterval: 30s
    buckets:
      - path: /headers
        bucket:
          maxTokens: 2
          tokensPerFill: 2
          fillInterval: 30s
      - path: /anything
        bucket:
          maxTokens: 2
          tokensPerFill: 2
          fillInterval: 30s
      - path: /ip
        bucket:
          maxTokens: 50
          tokensPerFill: 10
          fillInterval: 30s
```

EnvoyFilter

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: httpbin-local-rate-limit
  namespace: default
spec:
  workloadSelector:
    labels:
      app: httpbin
  configPatches:
    - applyTo: HTTP_FILTER
      match:
        context: SIDECAR_INBOUND
        listener:
          filterChain:
            filter:
              name: "envoy.filters.network.http_connection_manager"
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.http.local_ratelimit
          typed_config:
            "@type": type.googleapis.com/udpa.type.v1.TypedStruct
            type_url: type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
            value:
              stat_prefix: http_local_rate_limiter
    - applyTo: HTTP_ROUTE
      match:
        context: SIDECAR_INBOUND
      patch:
        operation: MERGE
        value:
          route:
            rate_limits:
              - actions:
                  - request_headers:
                      header_name: ":path"
                      descriptor_key: path
          typed_per_filter_config:
            envoy.filters.http.local_ratelimit:
              "@type": type.googleapis.com/udpa.type.v1.TypedStruct
              type_url: type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
              value:
                stat_prefix: rate_limit
                enable_x_ratelimit_headers: DRAFT_VERSION_03
                filter_enabled:
                  runtime_key: local_rate_limit_enabled
                  default_value:
                    numerator: 100
                    denominator: HUNDRED
                filter_enforced:
                  runtime_key: local_rate_limit_enforced
                  default_value:
                    numerator: 100
                    denominator: HUNDRED
                always_consume_default_token_bucket: false
                token_bucket:
                  max_tokens: 100
                  tokens_per_fill: 50
                  fill_interval: 30s
                descriptors:
                  - entries:
                      - key: path
                        value: /headers
                    token_bucket:
                      max_tokens: 2
                      tokens_per_fill: 2
                      fill_interval: 30s
                  - entries:
                      - key: path
                        value: /anything
                    token_bucket:
                      max_tokens: 2
                      tokens_per_fill: 2
                      fill_interval: 30s
                  - entries:
                      - key: path
                        value: /ip
                    token_bucket:
                      max_tokens: 50
                      tokens_per_fill: 10
                      fill_interval: 30s
```

#### Rate Limit by Request Header

The RateLimit CR allows configuring rate limits for specific headers. The **local.rateLimits** field is a list of
additional rate limit configurations that contain matching criteria such as headers.  
Each header must be configured with a name and value. If multiple headers are configured for a single bucket, the rate
limit is applied only if all specified headers are present in the request with the given values.

There is a limitation when it comes to header values for local rate limiting. Unlike global rate limiting,
the `header_request` [RateLimit Action](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-ratelimit-action-request-headers)
requires a static descriptor value for local rate limiting. This means that the header value and must be
defined in RateLimit CR. The header value doesn't support regex or wildcards (`*`), so the rate limit is applied only if
the
header value matches the configured value.

Example for EnvoyFilter creation based on the RateLimit CR configuration:

RateLimit CR

```yaml
apiVersion: ratelimit.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  name: httpbin-local-rate-limit
  namespace: default
spec:
  selectorLabels:
    app: httpbin
  local:
    defaultBucket:
      maxTokens: 1
      tokensPerFill: 1
      fillInterval: 60s
    buckets:
      - headers:
          x-client-type: external
          x-api-version: v1
        bucket:
          maxTokens: 10
          tokensPerFill: 10
          fillInterval: 60s
      - headers:
          x-client-type: external
          x-api-version: v2
        bucket:
          maxTokens: 15
          tokensPerFill: 15
          fillInterval: 60s
      - headers:
          x-client-type: internal
          x-api-version: v1
        bucket:
          maxTokens: 40
          tokensPerFill: 40
          fillInterval: 60s
      - headers:
          x-client-type: internal
          x-api-version: v2
        bucket:
          maxTokens: 60
          tokensPerFill: 60
          fillInterval: 60s
```

EnvoyFilter

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: httpbin-local-rate-limit
  namespace: default
spec:
  workloadSelector:
    labels:
      app: httpbin
  configPatches:
    - applyTo: HTTP_FILTER
      match:
        context: SIDECAR_INBOUND
        listener:
          filterChain:
            filter:
              name: "envoy.filters.network.http_connection_manager"
      patch:
        operation: INSERT_BEFORE
        value:
          name: envoy.filters.http.local_ratelimit
          typed_config:
            "@type": type.googleapis.com/udpa.type.v1.TypedStruct
            type_url: type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
            value:
              stat_prefix: http_local_rate_limiter
    - applyTo: HTTP_ROUTE
      match:
        context: SIDECAR_INBOUND
      patch:
        operation: MERGE
        value:
          route:
            rate_limits:
              - actions:
                  - request_headers:
                      header_name: "x-client-type"
                      descriptor_key: "x-client-type-key"
                  - request_headers:
                      header_name: "x-api-version"
                      descriptor_key: "x-api-version-key"
          typed_per_filter_config:
            envoy.filters.http.local_ratelimit:
              "@type": type.googleapis.com/udpa.type.v1.TypedStruct
              type_url: type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
              value:
                stat_prefix: rate_limit
                enable_x_ratelimit_headers: DRAFT_VERSION_03
                filter_enabled:
                  runtime_key: local_rate_limit_enabled
                  default_value:
                    numerator: 100
                    denominator: HUNDRED
                filter_enforced:
                  runtime_key: local_rate_limit_enforced
                  default_value:
                    numerator: 100
                    denominator: HUNDRED
                always_consume_default_token_bucket: false
                token_bucket:
                  max_tokens: 1
                  tokens_per_fill: 1
                  fill_interval: 60s
                descriptors:
                  - entries:
                      - key: x-client-type-key
                        value: external
                      - key: x-api-version-key
                        value: v1
                    token_bucket:
                      max_tokens: 10
                      tokens_per_fill: 10
                      fill_interval: 60s
                  - entries:
                      - key: x-client-type-key
                        value: external
                      - key: x-api-version-key
                        value: v2
                    token_bucket:
                      max_tokens: 15
                      tokens_per_fill: 15
                      fill_interval: 60s
                  - entries:
                      - key: x-client-type-key
                        value: internal
                      - key: x-api-version-key
                        value: v1
                    token_bucket:
                      max_tokens: 40
                      tokens_per_fill: 40
                      fill_interval: 60s
                  - entries:
                      - key: x-client-type-key
                        value: internal
                      - key: x-api-version-key
                        value: v2
                    token_bucket:
                      max_tokens: 60
                      tokens_per_fill: 60
                      fill_interval: 60s
```

### Enforcing Rate Limit

The [Envoy LocalRateLimit filter](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/local_ratelimit/v3/local_rate_limit.proto#local-rate-limit-proto)
allows to configure the enforcement for 0 - 100% percentage of the traffic by using the field **filter_enforced**.   
It is considered a good use case to allow to create a RateLimit CR without enforcing the rate limit. By following this
approach, it is possible to check the rate limit metrics to understand the impact on real traffic before blocking it.  
The decision is to hide the complexity of the Envoy enforcement configuration behind the optional boolean field
*spec.enforce* with the default value set to `true`.

### Rate Limit HTTP Response Headers

While enabling rate limit feature on the HTTP layer, there is a possibility to give insights to the client about the
current rate limiting state, for example, how many requests are left to be done before being rate limited, or how much time is
left until rate limit tokens are refilled.  
The decision is that the user can enable rate limit headers in the RateLimit CR by setting the boolean field
**spec.enableResponseHeaders**. Following security good practices, these headers are disabled by default to limit
internal information exposure.

### RateLimit CR Spec

| field                             | type                | description                                                                                                                                                                                                                                                                                                                                                                                        | required |
|-----------------------------------|---------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|
| selectorLabels                    | map<string, string> | Labels that specify the set of Pods to which the configuration applies.<br/> Each Pod can match only one RateLimit CR.<br/> The label search scope is limited to the namespace where the resource is located.                                                                                                                                                                                      | yes      |
| local                             | object              | Local rate limit configuration.                                                                                                                                                                                                                                                                                                                                                                    | yes      |
| local.defaultBucket               | object              | The default token bucket for rate limiting requests. <br/>If additional local buckets are configured in the same RateLimit CR, this bucket serves as a fallback for requests that don't match any other bucket's criteria. <br/>Each request consumes a single token. If a token is available, the request is allowed. If no tokens are available, the request is rejected with status code `429`. | yes      |
| local.defaultBucket.maxTokens     | int                 | The maximum tokens that the bucket can hold. This is also the number of tokens that the bucket initially contains.                                                                                                                                                                                                                                                                                 | yes      |
| local.defaultBucket.tokensPerFill | int                 | The number of tokens added to the bucket during each fill interval.                                                                                                                                                                                                                                                                                                                                | yes      |
| local.defaultBucket.fillInterval  | duration            | The fill interval that tokens are added to the bucket. <br/>During each fill interval, `tokensPerFill` are added to the bucket. The bucket will never contain more than `maxTokens` tokens. The `fillInterval` must be greater than or equal to 50ms to avoid excessive refills.                                                                                                                   | yes      |
| local.buckets                     | array               | List of additional rate limit buckets for requests. <br/>Each bucket must specify either a `path` or `headers`. <br/>For each request matching the bucket's criteria, a single token is consumed. If a token is available, the request is allowed. If no tokens are available, the request is rejected with status code `429`.                                                                     | no       |
| local.buckets.path                | string              | Specifies the path to be rate limited starting with `/`. <br/>For example, `/foo`.                                                                                                                                                                                                                                                                                                                     | no       |
| local.buckets.headers             | map<string, string> | Specifies the request headers to be rate limited. The key is the header name, and the value is the header value. All specified headers must be present in the request for this configuration to match. For example, `x-api-usage: BASIC`.                                                                                                                                                              | no       |                                                                                                                                                                                    
| local.buckets.maxTokens           | int                 | The maximum tokens that the bucket can hold. This is also the number of tokens that the bucket initially contains.                                                                                                                                                                                                                                                                                 | yes      |
| local.buckets.tokensPerFill       | int                 | The number of tokens added to the bucket during each fill interval.                                                                                                                                                                                                                                                                                                                                | yes      |
| local.buckets.fillInterval        | duration            | The fill interval that tokens are added to the bucket. <br/>During each fill interval, `tokensPerFill` are added to the bucket. The bucket will never contain more than `maxTokens` tokens. The `fillInterval` must be greater than or equal to 50ms to avoid excessive refills.                                                                                                                   | yes      |
| enableResponseHeaders             | boolean             | Enables **x-rate-limit** response headers. The default value is `false`.                                                                                                                                                                                                                                                                                                                           | no       |
| enforce                           | boolean             | Specifies whether the rate limit should be enforced. The default value is `true`.                                                                                                                                                                                                                                                                                                                            | no       |

### Usage Example

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: RateLimit
metadata:
  name: httpbin-local-rate-limit
  namespace: default
spec:
  selectorLabels:
    app: httpbin
  local:
    defaultBucket:
      maxTokens: 10
      tokensPerFill: 5
      fillInterval: 30s
    buckets:
      - path: /headers
        headers:
          x-api-version: v1
        bucket:
          maxTokens: 2
          tokensPerFill: 2
          fillInterval: 30s
      - path: /ip
        bucket:
          maxTokens: 20
          tokensPerFill: 10
          fillInterval: 30s
  enableResponseHeaders: true
```

## Consequences

A new controller for the new RateLimit CRD must be implemented as a part of the API Gateway module. The new RateLimit CR can
be used to set rate limits in the cluster's service mesh without having to worry about possible changes in the Istio
EnvoyFilter resources. Nevertheless, EnvoyFilter is complex and since its API is not stable yet, the current RateLimit
CR will be provided in `v1alpha1` version and might be changed in the future releases.    
The RateLimit CRD should be included in the blocking deletion strategy for APIGateway CR, since APIGateway managed resources
should not be uninstalled while RateLimit CRs exist in the cluster.
Requests beyond the allowed rate limit threshold will get the HTTP `429` response.

Similarly to the Istio based APIRules, the rate limit functionality requires the Istio service mesh to be present in the cluster.

In general, it should not be necessary for users to create resources in the `istio-system` namespace. However, for rate
limiting the Istio Ingress Gateway, the RateLimit CR must be created in the `istio-system` namespace.

With RateLimit CR supporting multiple rate limit configuration entries it is hard to create accurate metrics for each of
the rate limits. Per default, the metrics are disabled. If enabled, the metrics will be aggregated, and it will be
hard to distinguish which rate limit configuration is causing the rate limit to be hit.
