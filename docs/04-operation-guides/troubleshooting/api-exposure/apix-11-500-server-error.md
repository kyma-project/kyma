---
title: Cannot connect to a service exposed by an API Rule - 500 Internal Server Error.
---

## Symptom

Having an apirule under one host URL with wildcard /* and specific /headers paths with different handlers.

  ```bash
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: sample-apirule
      namespace: $NAMSEPSACE
    spec:
      gateway: kyma-system/kyma-gateway
      host: httpbin.$DOMAIN
      service:
        name: httpbin
        port: 8000
      rules:
        - path: /.*
          methods: ["GET"]
          accessStrategies:
            - handler: noop
        - path: /headers
          methods: ["GET"]
          accessStrategies:
            - handler: oauth2_introspection
              config:
                required_scope: ["read"]
  ```


When you try to reach your service, you get `500 Internal Server Error` in response.
  ```bash
  {"error":{"code":500,"status":"Internal Server Error","request":"e84400db-16b3-4818-9370-f10a6b4f3876","message":"An internal server error occurred, please contact the system administrator"}}
  ```

## Remedy

This issue was reported in the Ory Oathkeeper [project](https://github.com/ory/oathkeeper/issues/157). 