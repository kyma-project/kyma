This program translates old api (Api) to a new one (ApiRule)
Usage:
go run cmd/main.go --omit-apis-with-labels=migration/status

The command finds the legacy API objects and transforms it to the "APIRule" version.
Example api(s) that do work are below.
TODO: ensure the tool correctly reports all cases when api(s) can't be migrated automatically,


--------------------------------------------------------------------------------
apiVersion: gateway.kyma-project.io/v1alpha2
kind: Api
metadata:
    name: httpbin-api
spec:
    service:
      name: httpbin
      port: 8000
    hostname: httpbin.kyma.local
    authentication:
    - type: JWT
      jwt:
        issuer: https://dex.kyma.local
        jwksUri: http://dex-service.kyma-system.svc.cluster.local:5556/keys
        triggerRule:
          excludedPaths:
          - suffix: /favicon.ico
          - regex: /anything.+
--------------------------------------------------------------------------------
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: httpbin
spec:
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  service:
    name: httpbin
    port: 8000
    host: httpbin-new.kyma.local
  rules:
    - path: /favicon.ico
      methods: ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"]
      accessStrategies:
        - handler: allow
    - path: /anything.+
      methods: ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"]
      accessStrategies:
        - handler: allow
    - path: /.*
      methods: ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"]
      accessStrategies:
        - handler: jwt
          config:
            trusted_issuers:
              - "https://dex.kyma.local"
            jwks_urls:
              - "http://dex-service.kyma-system.svc.cluster.local:5556/keys"
--------------------------------------------------------------------------------
