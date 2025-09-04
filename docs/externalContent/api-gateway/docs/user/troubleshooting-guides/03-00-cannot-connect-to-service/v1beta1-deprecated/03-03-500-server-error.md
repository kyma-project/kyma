# 500 Internal Server Error

## Symptom

You have a deployed APIRule that looks similar to the following one: 

```bash
apiVersion: gateway.kyma-project.io/v1beta1
kind: APIRule
metadata:
  name: sample-apirule
  namespace: $NAMESPACE
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
The APIRule is configured under one host URL with the `/*` wildcard, the specific `/headers` path, and the same `GET` methods, which use different handlers.
When you try to reach your Service, you get the `500 Internal Server Error` response:

```bash
{"error":{"code":500,"status":"Internal Server Error","request":"e84400db-16b3-4818-9370-f10a6b4f3876","message":"An internal server error occurred, please contact the system administrator"}}
```

## Cause

Having multiple APIRules defined under the same host URL carries the risk of errors for specific paths due to the configuration overlap in Oathkeeper.
The root cause of the problem is the lack of support for the negative lookahead in the Golang language. For more information, see [the issue](https://github.com/ory/oathkeeper/issues/157) reported in the Ory Oathkeeper project.

## Solution

To resolve the issue, follow these guidelines:

- Set different hosts for different access strategies:

    ```bash
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: sample-apirule
      namespace: $NAMESPACE
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
      namespace: $NAMESPACE
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

- Set different methods for the specified paths:

    ```bash
    apiVersion: gateway.kyma-project.io/v1beta1
    kind: APIRule
    metadata:
      name: sample-apirule
      namespace: $NAMESPACE
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


When Oathkeeper throws the `503 Service Unavailable` or `502 Bad Gateway` responses, try to restart the Pod in order to resolve the issue. If you want to investigate what caused the error, follow these steps:

1. Check all Oathkeeper Pods:

    ```bash
    kubectl get pods -n kyma-system -l app.kubernetes.io/name=oathkeeper
    ```

2. Check if the load is heavy on the listed Pods:

    ```bash
   kubectl top pods -n kyma-system -l app.kubernetes.io/name=oathkeeper
    ```

3. Access the logs to check for other Oathkeeper errors:

    ```bash
    kubectl logs -n kyma-system -l app.kubernetes.io/name=oathkeeper  -c oathkeeper
    ```
