# API Gateway Migrator

The API Gateway Migrator translates the existing API objects (APIs) to new ones (APIRules).

## Usage:
To migrate APIs to APIRules, run:
```bash
go run cmd/main.go --label-blacklist=migration/status
```

To see all possible arguments, run:
```bash
go run cmd/main.go -h
```

The migration process randomizes the host to point to a non-existing location. This disables the existing APIs but does not remove the objects. To delete all migrated API objects, run:
```bash
kubectl delete apis -l migration/status=migrated --all-namespaces
```

## Development
### API migration example

>**NOTE:** For brevity, the examples do not include any metadata. For details about the role of the metadata in the migration process, see the [metadata](./#metadata) section.

This example shows a sample input API:
```yaml
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
```

This example shows an APIRule resulting from the migration:
```yaml
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
```

For more examples, see the [examples](./examples/) folder.
The folder includes an example of a [complex API](./examples/invalid.for.migration.input.yaml) object that the migration tool does **not** migrate automatically.
You can still migrate such objects manually using complex regular expressions for paths. [Here](./examples/invalid.for.migration.output.yaml) you can see an example of a manually created APIRule that corresponds to a complex API.

## Metadata

The migration process uses labels and annotations to describe the state of migrated objects.

After the migration, the following metadata is set on the migrated API object:

- The **migration/status**  label with the value `migrated`.
- The **migration/host** annotation with the original value of the API's **spec.hostname** parameter set before migration.

After the migration, the following metadata is set on the created APIRule object:

- The **migratedFrom** label with the value equal to the name of the migrated API.
- The **targetHost** annotation with the original value of the API's **spec.hostname** parameter set before migration.

## Rules for skipping API objects

During the migration, the tool can skip some API objects for the following reasons:
- API is already migrated
- API has an invalid status
- API labels match the configured **label-blacklist** parameter
- API is too complex for automatic translation

The API is considered too complex if it has more than one `authentication.jwt` element and every `jwt` object has different nested `excludedPaths`.
If the API has more than one `authentication.jwt` element, but all of such elements have the same nested `excludedPaths` (or none at all), the migration tool will translate it.
