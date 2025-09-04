# Issues When Creating an APIRule in Version v1beta1

## Symptom

When you create an APIRule, an instant validation error appears, or the APIRule custom resource (CR) has the `ERROR` status, for example:

```bash
kubectl get apirules.v1beta1.gateway.kyma-project.io httpbin

NAME      STATUS   HOST
httpbin   ERROR    httpbin.xxx.shoot.canary.k8s-hana.ondemand.com
```

The error may result in an inconsistent state of the APIRule resource in which Ory CR, Istio CR, or both are missing. Your Service then cannot be properly exposed.
To check the error message of the APIRule resource, run:

```bash
kubectl get apirules.v1beta1.gateway.kyma-project.io -n <namespace> <api-rule-name> -o=jsonpath='{.status.APIRuleStatus}'
```
---
## JWT Handler's **trusted_issuers** Configuration Is Missing
### Cause

The following APIRule is missing the **trusted_issuers** configuration for the JWT handler:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: jwt
```

If your APIRule is missing the **trusted_issuers** configuration for the JWT handler, the following `APIRuleStatus` error appears:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.rules[0].accessStrategies[0].config\": supplied config cannot be empty"}
```

### Remedy

Add JWT configuration for the **trusted_issuers** or ``. Here's an example of a valid configuration:

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
## Invalid **trusted_issuers** for the JWT Handler
### Cause

Here's an example of an APIRule with the **trusted_issuers** URL configured:

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

If the **trusted_issuers** URL is an unsecured HTTP URL, or the **trusted_issuers** URL is not valid, you get an instant error, and the APIRule resource is not created:

```
The APIRule "httpbin" is invalid: spec.rules[0].accessStrategies[0].config.trusted_issuers[0]: Invalid value: "some-url": spec.rules[0].accessStrategies[0].config.trusted_issuers[0] in body should match '^(https://|file://).*$'
```

### Remedy

The JWT **trusted-issuers** must be a valid HTTPS URL, for example:

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
## Unsupported Handlers' Combination
### Cause

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

### Remedy

Decide on one configuration you want to use. You can either `allow` access to the specific path or restrict it using the JWT security token. Defining both configuration methods on the same path is not allowed.

---
## Occupied Host
### Cause

The following APIRules use the same host:

```yaml
spec:
  ...
spec:
  host: httpbin.xxx.shoot.canary.k8s-hana.ondemand.com
```

If your APIRule specifies a host that is already used by another APIRule or Virtual Service, the following `APIRuleStatus` error appears:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.host\": This host is occupied by another Virtual Service"}
```

### Remedy

Use a different host for the second APIRule, for example:

```yaml
spec:
  ...
  host: httpbin-new.xxx.shoot.canary.k8s-hana.ondemand.com
```

---
## Configuration of `noop`, `allow`, and `no_auth` Handlers 
### Cause

In the following APIRule, the `noop` handler has the **trusted-issuers** field configured:

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

If your APIRule uses either the `noop`, `allow`, or `no_auth` handler and has some further handler's configuration defined, you get the following `APIRuleStatus` error:

```
{"code":"ERROR","desc":"Validation error: Attribute \".spec.rules[0].accessStrategies[0].config\": strategy: noop does not support configuration"}
```

### Remedy

Use the `noop`, `allow`, and `no_auth` handlers without any further configuration, for example:

```yaml
spec:
  ...
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: noop
```