## Overview

This program translates old api objects (Api) to a new one (ApiRule)
Usage:
go run cmd/main.go --label-blacklist=migration/status

To see all possible arguments, use the following:
go run cmd/main.go -h

After migration, old Api objects are disabled - the host is randomized to point to a non-existing location.
Objects themselves are not deleted. Users can delete all migrated Api objects with the following command:
`kubectl delete apis -l migration/status=migrated --all-namespaces`

## Api migration example

Sample api and migration results are are shown below.

#### Input Api
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

#### Output ApiRule
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

For more examples take a look into ./examples folder.
You can find there also an example of a complicated api object that will NOT be automatically migrated by this tool: ./examples/invalid.for.migration.input.yaml
Such objects can still be migrated manually using rather complex regular expressions for paths. You can see an example of manually created ApiRule that corresponds to a complicated api it in the file: ./examples/invalid.for.migration.output.yaml.


## The rules for skipping api objects

The tool skips api objects for the following reasons:
- api is already migrated
- api has invalid status
- api labels are matching configured "label-blacklist" parameter
- api is too complex for automatic translation

The api is considered too complex if it has more than one `authentication.jwt` element and every `jwt` object has different nested `excludedPaths`.
If the api has more than one `authentication.jwt` elements but all of them have the same nested `excludedPaths` (or none at all), it will be translated.
