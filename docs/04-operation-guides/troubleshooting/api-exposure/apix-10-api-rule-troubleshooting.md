---
title: Issues when creating an APIRule - Various reasons
---

## Symptom

When you create an APIRule, an instant validation error appears, or the APIRule custom resource (CR) has the `ERROR` status, for example:

```bash
kubectl get apirule httpbin

NAME      STATUS   HOST
httpbin   ERROR    httpbin.xxx.shoot.canary.k8s-hana.ondemand.com
```

The error may result in an inconsistent state of the APIRule resource in which Ory CR, Istio CR, or both are missing. Your service then cannot be properly exposed.
To check the error message of the APIRule resource, run:

```bash
kubectl get apirule -n <namespace> <api-rule-name> -o=jsonpath='{.status.APIRuleStatus}'
```
---
## JWT handler's `trusted_issuers` configuration is missing
#### Cause

The following APIRule is missing the `trusted_issuers` configuration for the JWT handler:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: jwt
```

If your APIRule is missing the `trusted_issuers` configuration for the JWT handler, the following `APIRuleStatus` error appears:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.rules[0].accessStrategies[0].config\": supplied config cannot be empty"}
```

#### Remedy

Add JWT configuration for the `trusted_issuers` or ``. Here's an example of a valid configuration:

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
## Invalid `trusted_issuer` for the JWT handler
#### Cause

Here's an example of the APIRule with a `trusted_issuer` URL configured:

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
            trusted_issuers: ["http://unsecured.or.not.valid.url"]
EOF
```

If the `trusted_issuer` URL is an unsecured HTTP URL, or the `trusted_issuer` URL is not valid, you get an instant error, and the APIRule resource is not created:

```
The APIRule "httpbin" is invalid: spec.rules[0].accessStrategies[0].config.trusted_issuers[0]: Invalid value: "some-url": spec.rules[0].accessStrategies[0].config.trusted_issuers[0] in body should match '^(https://|file://).*$'
```

#### Remedy

The JWT `trusted-issuer` must be a valid HTTPS URL, for example:

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
## Unsupported handlers' combination
#### Cause

The following APIRule has both `allow` and `jwt` handlers defined on the same path:

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

The handlers' combination in the above example is not supported. If an APIRule has an unsupported handlers' combination defined **on the same path**, the following `APIRuleStatus` error appears:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.rules[0].accessStrategies.accessStrategies[0].handler\": allow access strategy is not allowed in combination with other access strategies"}
```

#### Remedy

You should decide to `allow` access to the specific path or restrict it via `jwt` security token. Using both at the same time is not allowed.

---
## Issue - Service in APIRule is blocklisted

If your APIRule specifies a service that is blocklisted, e.g.:

```yaml
spec:
  ...
  service:
    name: istio-ingressgateway
    namespace: istio-system
```

You will get the following `APIRuleStatus` error:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.service.name\": Service istio-ingressgateway in namespace istio-system is blocklisted"}
```

## Remedy

Please refer to the technical reference documentation for the [blocklisted services in API Gateway](https://kyma-project.io/docs/kyma/latest/05-technical-reference/apix-03-blacklisted-services/)

Also you may check the [serviceBlockList](https://github.com/kyma-project/kyma/blob/main/resources/api-gateway/values.yaml) defaults definition in Kyma.

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
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: noop
```
