# Proposal for APIRule API CORS Headers Configuration in IstioVirtual Service

## API Proposal

In Istio VirtualService, you can configure CORS using the following parameters:
- **AllowHeaders** - defines a list of HTTP headers that you can use when requesting a resource. The contents of this list are serialized into the **Access-Control-Allow-Headers** header.
- **AllowMethods** - defines a list of HTTP methods allowed to access the resource. The contents of this list are serialized into the **Access-Control-Allow-Methods** header.
- **AllowOrigins** - specifies string patterns that match allowed origins. An origin is allowed if any of the string matchers find a match. If a match is found, the outgoing **Access-Control-Allow-Origin** header is set to the origin provided by the client. The value must be of the type [StringMatch](https://istio.io/latest/docs/reference/config/networking/virtual-service/#StringMatch).
- **AllowCredentials** - specifies whether the caller is allowed to send the actual request (not the preflight) using credentials. It translates into the **Access-Control-Allow-Credentials** header. The value can be either `true` or `false`.
- **ExposeHeaders** - defines a list of HTTP headers that browsers are allowed to access. The contents of this list are serialized into the **Access-Control-Expose-Headers** header.
- **MaxAge** - determines the duration for which the results of a preflight request can be stored in cache. The parameter translates into the **Access-Control-Max-Age** header.

The chosen configuration must allow for the exposure of all the listed parameters. The following structure is capable of storing this information:
```go
type CorsPolicy struct {
	AllowHeaders     []string               `json:"allowHeaders,omitempty"`
	AllowMethods     []string               `json:"allowMethods,omitempty"`
	AllowOrigins     []*v1beta1.StringMatch `json:"allowOrigins,omitempty"`
	AllowCredentials bool                   `json:"allowCredentials"`
	ExposeHeaders    []string               `json:"exposeHeaders,omitempty"`
	MaxAge           *time.Duration         `json:"maxAge,omitempty"`
}
```

## Security Considerations

### Default Values

In the most secure scenario, CORS should be configured not to respond with any of the **Access-Control** headers. However, for backward compatibility with the current implementation, it is important to note that the current configuration for all APIRules is as follows:
```yaml
CorsAllowOrigins: "regex:.*"
CorsAllowMethods: "GET,POST,PUT,DELETE,PATCH"
CorsAllowHeaders: "Authorization,Content-Type,*"
```

**Decision**
The default values should be empty to ensure the secure by default configuration. The transition to that default should be gradual, with the current CORS configuration staying in place for now.

### CORS Headers Sanitization

If a workload provides its own CORS headers, Istio Ingress Gateway does not sanitize or change the CORS headers unless the request origin matches one of the origins specified in the **AllowOrigins** configuration of the VirtualService. This can pose a security risk because it may be unexpected for the server response to contain different headers than those defined in the APIRule.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: oauth2-test-ckxrj
  namespace: default
spec:
  gateways:
  - kyma-system/kyma-gateway
  hosts:
  - hello.local.kyma.dev
  http:
  - corsPolicy:
      allowHeaders:
      - Authorization
      - Content-Type
      - '*'
      allowMethods:
      - GET
      - POST
      - PUT
      - DELETE
      - PATCH
      allowOrigins:
      - exact: https://test.com
    headers:
      request:
        set:
          x-forwarded-host: hello.local.kyma.dev
    match:
    - uri:
        regex: /.*
    route:
    - destination:
        host: helloworld.default.svc.cluster.local
        port:
          number: 5000
      weight: 100
    timeout: 180s

```

**Decision**
APIRule must be the only source of truth and disregard the upstream response headers. If the CORS configuration is empty, the default configuration specified in [Default values](#default-values) should be applied.
