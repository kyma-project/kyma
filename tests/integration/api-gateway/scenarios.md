# Expose without access strategy next update to JWT/Oatuh2 access strategy (whole service)

```yaml
---
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: noop2jwt-strategy
spec:
  service:
    host: httpbin1.kyma.local
    name: httpbin
    port: 8000
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: noop
        # - handler: jwt
        #   config:
        #     trusted_issuers: ["http://dex.kyma.local"]
        #     required_scope: []
      mutators: []
---
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: noop2jwt-strategy
spec:
  service:
    host: httpbin1.kyma.local
    name: httpbin
    port: 8000
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        # - handler: noop
        - handler: jwt
          config:
            trusted_issuers: ["http://dex.kyma.local"]
            required_scope: []
      mutators: []
```

# Expose without access strategy next update to JWT/Oauth2 access strategy (on path)

```yaml
---
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: noop2jwt-strategy-multiple-paths
spec:
  service:
    host: httpbin1.kyma.local
    name: httpbin
    port: 8000
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
    - path: /cookies
      methods: ["GET"]
      accessStrategies:
        - handler: noop
      mutators: []
    - path: /headers
      methods: ["GET"]
      accessStrategies:
        - handler: noop
        # - handler: jwt
        #   config:
        #     trusted_issuers: ["http://dex.kyma.local"]
        #     required_scope: []
      mutators: []
---
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: noop2jwt-strategy-multiple-paths
spec:
  service:
    host: httpbin1.kyma.local
    name: httpbin
    port: 8000
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
    - path: /cookies
      methods: ["GET"]
      accessStrategies:
        - handler: noop
      mutators: []
    - path: /headers
      methods: ["GET"]
      accessStrategies:
        # - handler: noop
        - handler: jwt
          config:
            trusted_issuers: ["http://dex.kyma.local"]
            required_scope: []
      mutators: []
```

# Expose with JWT/Oauth2 access strategy next update to give plan access

```yaml
---
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: jwt2noop-strategy
spec:
  service:
    host: httpbin1.kyma.local
    name: httpbin
    port: 8000
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
    - path: /headers
      methods: ["GET"]
      accessStrategies:
        # - handler: noop
        - handler: jwt
          config:
            trusted_issuers: ["http://dex.kyma.local"]
            required_scope: []
      mutators: []
---
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: jwt2noop-strategy
spec:
  service:
    host: httpbin1.kyma.local
    name: httpbin
    port: 8000
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
    - path: /headers
      methods: ["GET"]
      accessStrategies:
        - handler: noop
        # - handler: jwt
        #   config:
        #     trusted_issuers: ["http://dex.kyma.local"]
        #     required_scope: []
      mutators: []
```

# Create ApiRule next delete it

