# CORS Headers Configuration for APIRules in Version `v1`

## Status
Accepted

## Context

In preparation for exposing CORS configuration in APIRules, we must determine the default configuration that is used when no explicit configuration is provided. As of the creation day of ADR, all APIRules in version `v1beta1` use the following CORS configuration:

```yaml
Access-Control-Allow-Origins: "*"
Access-Control-Allow-Methods: "GET,POST,PUT,DELETE,PATCH"
Access-Control-Allow-Headers: "Authorization,Content-Type,*"
```

## Decision

- The configuration for the current APIRule in version `v1beta1` will remain unchanged.
- Starting from APIRule version `v1`, the default configuration for CORS headers will be modified. It will no longer include any headers unless they are explicitly specified in the APIRule. The APIRule will become the sole source of truth for CORS configuration.

## Consequences

- The change in default CORS configuration from version `v1beta1` to the new version `v1` is a breaking change. Therefore, it must be explicitly stated and communicated.
- As part of the migration from `v1beta1` to `v1`, customers will be required to manually configure CORS.