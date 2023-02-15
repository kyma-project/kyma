---
title: Cannot connect to a service exposed by an API Rule - 500 Internal Server Error.
---

## Symptom

You have a deployed APIRule that looks similar to the following one: 

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
It is configured under one host URL with wildcard /* and specific /headers paths and the same GET method with different handlers.
When you try to reach your service, you get `500 Internal Server Error` in response.
  ```bash
  {"error":{"code":500,"status":"Internal Server Error","request":"e84400db-16b3-4818-9370-f10a6b4f3876","message":"An internal server error occurred, please contact the system administrator"}}
  ```

## Remedy

This issue was reported in the Ory Oathkeeper [project](https://github.com/ory/oathkeeper/issues/157).
Having multiple rules for one host URL is causing errors for specific paths due to the configuration overlap in the Oathkeeper.
This is based on lack of support of negative lookahead in the Golang language.

To resolve this issue, you can try those guidelines:

- Set the different hosts for different access startegies:

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
  ```

  ```bash
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: sample-apirule-secured
      namespace: $NAMSEPSACE
    spec:
      gateway: kyma-system/kyma-gateway
      host: httpbin-secured.$DOMAIN
      service:
        name: httpbin
        port: 8000
      rules:
        - path: /headers
          methods: ["GET"]
          accessStrategies:
            - handler: oauth2_introspection
              config:
                required_scope: ["read"]
  ```

- Set the different methods for specified paths:

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
          methods: ["POST"]
          accessStrategies:
            - handler: noop
        - path: /headers
          methods: ["GET"]
          accessStrategies:
            - handler: oauth2_introspection
              config:
                required_scope: ["read"]
  ```


Sometimes Oathkeeper can throw `503 Service Unavailable` or `502 Bad Gateway` responses. While a simple restart of the Pod could resolve this issue, you might want to verify it and check what issue is causing this. To do so, follow these steps:

1. Check all Oathkeeper Pods:

    ```bash
    kubectl get pods -n kyma-system -l app.kubernetes.io/name=oathkeeper
    ```

2. Check if the load is heavy on those Pods:

    ```bash
   kubectl top pods -n kyma-system -l app.kubernetes.io/name=oathkeeper
   ```

3. Access logs to check for any other errors in the Oathkeeper:

    ```bash
    kubectl logs -n kyma-system -l app.kubernetes.io/name=oathkeeper  -c oathkeeper
   ```
