# Issues When Creating an APIRule Custom Resource in Version v2alpha1

## Symptom

When you create an APIRule custom resource (CR), an instant validation error appears, or the APIRule CR has the `ERROR` status, for example:

```bash
kubectl get apirules httpbin

NAME      STATUS   HOST
httpbin   ERROR    httpbin.xxx.shoot.canary.k8s-hana.ondemand.com
```

The error may signify that your APIRule CR is in an inconsistent state and the service cannot be properly exposed.
To check the error message of the APIRule CR, run:


```bash
kubectl get apirules -n <namespace> <api-rule-name> -o=jsonpath='{.status.description}'
```

---
## The **issuer** Configuration of **jwt** Access Strategy Is Missing
### Cause

The following APIRule is missing the **issuer** configuration for the **jwt** access strategy:

```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  ...
spec:
  ...
  rules:
    - path: /*
      jwt:
```

If your APIRule is missing the **issuer** configuration for the **jwt** access strategy, the following error appears:

```
{"code":"ERROR","description":"Validation error: Attribute \".spec.rules[0].jwt\": supplied config cannot be empty"}
```

### Solution

Add JWT configuration for the **issuer** or ``. Here's an example of a valid configuration:

```yaml
spec:
  ...
  rules:
    - path: /*
      methods: ["GET"]
      jwt:
        authentications:
          - issuer: "https://dev.kyma.local"
            jwksUri: "https://example.com/.well-known/jwks.json"
```

## Invalid **issuer** for the **jwt** Access Strategy
### Cause

Here's an example of the APIRule CR with an invalid **issuer** URL configured:

```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  ...
spec:
  ...
  rules:
    - path: /*
      jwt:
        authentications:
          - issuer: ://unsecured.or.not.valid.url
            jwksUri: https://example.com/.well-known/jwks.json
```

If the **issuer** contains `:`, it must be a valid URI. Otherwise, you get the following error, and the APIRule CR is not created:

```
The APIRule "httpbin" is invalid: .spec.rules[0].jwt.authentications[0].issuer: value is empty or not a valid URI
```

### Solution

The JWT **issuer** must not be empty and must be a valid URI, for example:

```yaml
apiVersion: gateway.kyma-project.io/v2alpha1
kind: APIRule
metadata:
  ...
spec:
  ...
  rules:
    - path: /*
      jwt:
        authentications:
          - issuer: https://dev.kyma.local
            jwksUri: https://dev.kyma.local/.well-known/jwks.json
```

---
## Both **noAuth** and **jwt** Access Strategies Defined on the Same Path
### Cause

The following APIRule CR has both **noAuth** and **jwt** access strategies defined on the same path:

```yaml
spec:
  ...
  rules:
    - path: /*
      noAuth: true
      jwt:
        authentications:
          - issuer: https://dev.kyma.local
            jwksUri: https://dev.kyma.local/.well-known/jwks.json
```

If you set the **noAuth** access strategy to `true` and define the **jwt** configuration on the same path, you get the following error:

```
{"code":"ERROR","description":"Validation error: Attribute \".spec.rules[0].noAuth\": noAuth access strategy is not allowed in combination with other access strategies"}
```

### Solution

Decide on one configuration you want to use. You can either use **noAuth** access to the specific path or restrict it using a JWT security token.

---
## Occupied Host
### Cause

The following APIRule CRs use the same host:

```yaml
spec:
  ...
spec:
  host: httpbin.xxx.shoot.canary.k8s-hana.ondemand.com
```

If your APIRule CR specifies a host that is already used by another APIRule or Virtual Service, the following error appears:

```
{"code":"ERROR","description":"Validation error: Attribute \".spec.host\": This host is occupied by another Virtual Service"}
```

### Solution

Use a different host for the second APIRule CR, for example:

```yaml
spec:
  ...
  host: httpbin-new.xxx.shoot.canary.k8s-hana.ondemand.com
```
