# Request Modifiers in APIRule

## Status

Accepted

## Context

Request modifiers are introduced to the APIRule. They can add cookies or headers to the request. This behaviour is similar to [Istio JWT mutators](../../user/custom-resources/apirule/v1beta1-deprecated/04-40-apirule-mutators.md) in the previous APIRule version.

## Decision

APIRule has new fields that can be used to add cookies or headers to the request. See the following table:

| Field                     | Description                                                                                                 |
|---------------------------|-------------------------------------------------------------------------------------------------------------|
| **rules.request**         | Defines request modification rules, which are applied before forwarding the request to the target workload. |
| **rules.request.cookies** | Specifies a list of cookie key-value pairs, that are forwarded inside the **Cookie** header.                |
| **rules.request.headers** | Specifies a list of header key-value pairs that are forwarded as header=value to the target workload.       |

## Examples

APIRule with cookie modifiers:
```
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: httpbin
  namespace: somenamespace
spec:
  gateway: somenamespace/gateway
  hosts:
    - httpbin.example.com
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      noAuth: true
      request:
        cookies:
          cookie-name: cookie-value
```

APIRule with header modifiers:
```
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  name: httpbin
  namespace: somenamespace
spec:
  gateway: somenamespace/gateway
  hosts:
    - httpbin.example.com
  service:
    name: httpbin
    port: 8000
  rules:
    - path: /headers
      methods: ["GET"]
      noAuth: true
      request:
        headers:
          header-name: header-value
```
