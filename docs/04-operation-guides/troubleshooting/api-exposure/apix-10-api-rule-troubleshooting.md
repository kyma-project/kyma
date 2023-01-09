---
title: Issues when creating an APIRule - Various reasons
---

## Symptom

When you create an APIRule, you get an instant validation error or your APIRule custom resource (CR) have `ERROR` status, e.g.

```bash
kubectl get apirule httpbin

NAME      STATUS   HOST
httpbin   ERROR    httpbin.xxx.shoot.canary.k8s-hana.ondemand.com
```

This may result in an inconsistent state with missing Ory and/or Istio CRs. Your service won't be properly exposed. There are various reasons for it.
You can check the error message of the APIRule resource, run:

```bash
kubectl get apirule -n <namespace> <api-rule-name> -o=jsonpath='{.status.APIRuleStatus}'
```
---
## Issue - Missing handler configuration

If your APIRule is missing required configuration, e.g. for `jwt` handler:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: jwt
```

You will get the following `APIRuleStatus` error:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.rules[0].accessStrategies[0].config\": supplied config cannot be empty"}
```

## Remedy

Supply `jwt` configuration for `trusted_issuers` or ``, e.g.:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers: ["https://dev.kyma.local"]
```
---
## Issue - Invalid trusted_issuer for `jwt` handler

If your APIRule has unsecured `http` URL trusted_issuer or the trusted_issuer is not a valid URL e.g. :

```bash
cat <<EOF | kubectl apply -f -
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  ...
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers: ["some-url"]
EOF
```

You will get an instant error and APIRule resource won't be created:

```
The APIRule "httpbin" is invalid: spec.rules[0].accessStrategies[0].config.trusted_issuers[0]: Invalid value: "some-url": spec.rules[0].accessStrategies[0].config.trusted_issuers[0] in body should match '^(https://|file://).*$'
```

## Remedy

JWT trusted issuer must be a valid `https` URL, e.g.:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers: ["https://dev.kyma.local"]
```
---
## Issue - Unsupported handler combination

If your APIRule is having unsupported handler combination on the **same** path, e.g. `allow` and `jwt` handlers:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: allow
        - handler: jwt
          config:
            trusted_issuers: ["https://dev.kyma.local"]
```

You will get the following `APIRuleStatus` error:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.rules[0].accessStrategies.accessStrategies[0].handler\": allow access strategy is not allowed in combination with other access strategies"}
```

## Remedy

You should decide to `allow` access to the specific path or restrict it via `jwt` security token. Both at the same time is not allowed.

---
## Issue - Using Istio `jwt` handler configuration for Ory `jwt` handler

If your APIRule is having Istio `jwt` handler configuration but you use Ory as `jwt` handler, e.g.:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: jwt
          config:
            authentications:
            - issuer: "https://example.com"
              jwksUri: "https://example.com/.well-known/jwks.json"
```

and API Gateway ConfigMap is having `ory` for `jwtHandler`, run:

```bash
kubectl get configmap/api-gateway-config -n kyma-system -o=jsonpath='{.data.api-gateway-config}'

jwtHandler: "ory"
```

You will get the following `APIRuleStatus` error:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.rules[0].accessStrategies[0].config.authentications\": Configuration for authentications is not supported with Ory handler"}
```

## Remedy

- If you want to use Istio `jwt` handler you should switch to it in the `api-gateway-config` ConfigMap, run:

```bash
kubectl patch configmap/api-gateway-config -n kyma-system --type merge -p '{"data":{"api-gateway-config":"jwtHandler: istio"}}'
```

- If you want to use Ory `jwt` handler you should use correct configuration, e.g.:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers: ["https://example.com"]
            jwks_urls: ["https://example.com/.well-known/jwks.json"]
```

You should always refer to the technical reference documentation for the [APIRule CR](https://kyma-project.io/docs/kyma/latest/05-technical-reference/00-custom-resources/apix-01-apirule/)

---
## Issue - Using Ory `jwt` handler configuration for Istio `jwt` handler

If your APIRule is having Ory `jwt` handler configuration but you use Istio as `jwt` handler, e.g.:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers: ["https://example.com"]
            jwks_urls: ["https://example.com/.well-known/jwks.json"]
```

and API Gateway ConfigMap is having `istio` for `jwtHandler`, run:

```bash
kubectl get configmap/api-gateway-config -n kyma-system -o=jsonpath='{.data.api-gateway-config}'

jwtHandler: "istio"
```

You will get the following `APIRuleStatus` error:

```
{"code":"ERROR","desc":"Multiple validation errors: \nAttribute \".spec.rules[0].accessStrategies[0].config.jwks_urls\": Configuration for jwks_urls is not supported with Istio handler\nAttribute \".spec.rules[0].accessStrategies[0].config.trusted_issuers\": Configuration for trusted_issuers is not supported with Istio handler"}
```

## Remedy

- If you want to use Ory `jwt` handler you should switch to it in the `api-gateway-config` ConfigMap, run:

```bash
kubectl patch configmap/api-gateway-config -n kyma-system --type merge -p '{"data":{"api-gateway-config":"jwtHandler: ory"}}'
```

- If you want to use Istio `jwt` handler you should use correct configuration, e.g.:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: jwt
          config:
            authentications:
            - issuer: "https://example.com"
              jwksUri: "https://example.com/.well-known/jwks.json"
```

You should always refer to the technical reference documentation for the [APIRule CR](https://kyma-project.io/docs/kyma/latest/05-technical-reference/00-custom-resources/apix-01-apirule/)

---
## Issue - Service in APIRule is blocklisted

If your APIRule specifies a service that is blocklisted, e.g.:

```yaml
spec:
  ...
  service:
    name: apiserver-proxy
    namespace: kyma-system
```

You will get the following `APIRuleStatus` error:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.service.name\": Service apiserver-proxy in namespace kyma-system is blocklisted"}
```

## Remedy

Please refer to the technical reference documentation for the [blocklisted services in API Gateway](https://kyma-project.io/docs/kyma/latest/05-technical-reference/apix-03-blacklisted-services/)

---
## Issue - Host already occupied

If your APIRule specifies a host that is already used in another APIRule or Virtual Service, e.g. having two APIRules specifying same host:

```yaml
spec:
  ...
spec:
  host: httpbin.xxx.shoot.canary.k8s-hana.ondemand.com
```

You will get the following `APIRuleStatus` error:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.host\": This host is occupied by another Virtual Service"}
```

## Remedy

You should use different host for the second APIRule, e.g.:

```yaml
spec:
  ...
  host: httpbin-new.xxx.shoot.canary.k8s-hana.ondemand.com
```

---
## Issue - noop or allow handlers 
with configuration

If your APIRule have a `noop` or `allow` handler and specifies some configuration, e.g.:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: noop
          config:
            trusted_issuers: ["https://dex.kyma.local"]
```

You will get the following `APIRuleStatus` error:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.rules[0].accessStrategies[0].config\": strategy: noop does not support configuration"}
```

## Remedy

You should use `noop` and `allow` handlers without any configurations, e.g.

```yaml
```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: noop
```
