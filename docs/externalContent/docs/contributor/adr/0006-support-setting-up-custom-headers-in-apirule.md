# Support Setting Up Custom Headers in an APIRule

## Status

Proposed

## Context

Users request a feature for APIRule to support setting up custom headers in the CR based on the VirtualService [Headers section](https://istio.io/latest/docs/reference/config/networking/virtual-service/#Headers). Users use application configuration where upstream connectivity requires some values to be passed as headers in a request. See discussion in this [GitHub issue](https://github.com/kyma-project/api-gateway/issues/808#issuecomment-1959121650).

## Decision

1. APIRule definition is extended with the `headers` section.
2. The `headers` section lets the user manipulate headers on the request and response level.
3. The `headers` section lets the user set additional labels that will be present in a request or response.
4. If the user-provided key in APIRule is already present in the request/response, the value is overridden with user-provided data.


### APIRule Extension
```yaml
kind: APIRule
metadata:
    name: rule
    namespace: default
spec:
    gateway: kyma-system/kyma-gateway
    headers:
        request:
            set:
                X-CLIENT-SSL-CN: "%DOWNSTREAM_PEER_SUBJECT%"
                X-CLIENT-SSL-SAN: "%DOWNSTREAM_PEER_URI_SAN%"
                X-CLIENT-SSL-ISSUER: "%DOWNSTREAM_PEER_ISSUER%"
(...)
```

The `APIRuleSpec` struct is extended with a variable of type `HeadersConfig`:
```go
type APIRuleSpec {
    // Headers contains information about header manipulation in APIRule
    Headers HeadersConfig
}

type HeadersConfig struct {
    // Request defines header manipulation options for all requests passing through an APIRule
    Request HeaderOptions
    // Response defines header manipulation options for all responses passing through an APIRule
    Response HeaderOptions
}

type HeaderOptions {
    // Set contains key/value pair map containing headers that are set in APIRule
    Set map[string]string
}
```
## Consequences

The user is allowed to set up custom headers in requests and responses that go through APIRule and its managed resources.