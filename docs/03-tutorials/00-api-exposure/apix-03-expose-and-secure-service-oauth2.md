---
title: Expose and secure a service with OAuth2
---

This tutorial shows how to expose and secure services or Functions using API Gateway Controller. The controller reacts to an instance of the API Rule custom resource (CR) and creates an Istio Virtual Service and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR. To interact with the secured services, the tutorial uses an OAuth2 client registered through the Hydra Maester controller.

It may be a follow-up to the [Use a custom domain to expose a service](./apix-01-own-domain.md) tutorial.

## Prerequisites

This tutorial is based on a sample HttpBin service deployment and a sample Function. To deploy or create one of those, follow the [Deploy a service](./apix-02-deploy-service.md) tutorial.

## Register an OAuth2 client and get tokens

1. Export your client as an environment variable:

  ```shell
  export CLIENT_NAME={YOUR_CLIENT_NAME}
  ```

2. Create an OAuth2 client with `read` and `write` scopes. Run:

  ```shell
  cat <<EOF | kubectl apply -f -
  apiVersion: hydra.ory.sh/v1alpha1
  kind: OAuth2Client
  metadata:
    name: $CLIENT_NAME
    namespace: $NAMESPACE
  spec:
    grantTypes:
      - "client_credentials"
    scope: "read write"
    secretName: $CLIENT_NAME
  EOF
  ```

3. Export the credentials of the created client as environment variables. Run:

  ```shell
  export CLIENT_ID="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_id}' | base64 --decode)"
  export CLIENT_SECRET="$(kubectl get secret -n $CLIENT_NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_secret}' | base64 --decode)"
  ```

4. Encode your client credentials and export them as an environment variable:

  ```shell
  export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
  ```

5. Get tokens to interact with secured resources using client credentials flow:

   <div tabs>
     <details>
     <summary>
     Token with "read" scope
     </summary>

     1. Get the token:

         ```shell
         curl -ik -X POST "https://oauth2.$DOMAIN/oauth2/token" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=read"
         ```

     2. Export the issued token as an environment variable:

         ```shell
         export ACCESS_TOKEN_READ={ISSUED_READ_TOKEN}
         ```

     </details>
     <details>
     <summary>
     Token with "write" scope
     </summary>

     1. Get the token:

         ```shell
         curl -ik -X POST "https://oauth2.$DOMAIN/oauth2/token" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=write"
         ```

     2. Export the issued token as an environment variable:

         ```shell
         export ACCESS_TOKEN_WRITE={ISSUED_WRITE_TOKEN}
         ```

      </details>
   </div>

## Expose, and secure the sample resources

Follow the instructions in the tabs to expose an instance of the HttpBin service or a sample Function, and secure them with Oauth2 scopes.

<div tabs>

  <details>
  <summary>
  HttpBin - secure endpoints of a service
  </summary>

1. Expose the service and secure it by creating an API Rule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

  ```shell
  cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: httpbin
    namespace: $NAMESPACE
  spec:
    gateway: namespace-name/httpbin-gateway #The value corresponds to the Gateway CR you created. 
    service:
      name: httpbin
      port: 8000
      host: httpbin.$DOMAIN
    rules:
      - path: /.*
        methods: ["GET"]
        accessStrategies:
          - handler: oauth2_introspection
            config:
              required_scope: ["read"]
      - path: /post
        methods: ["POST"]
        accessStrategies:
          - handler: oauth2_introspection
            config:
              required_scope: ["write"]
  EOF
  ```

>**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

The exposed service requires tokens with "read" scope for `GET` requests in the entire service and tokens with "write" scope for `POST` requests to the `/post` endpoint of the service.

  </details>

  <details>
  <summary>
  Secure a Function
  </summary>

1. Expose the service and secure it by creating an API Rule CR in your Namespace. If you don't want to use your custom domain but a Kyma domain, use the following Kyma Gateway: `kyma-system/kyma-gateway`. Run:

  ```shell
  cat <<EOF | kubectl apply -f -
  apiVersion: gateway.kyma-project.io/v1alpha1
  kind: APIRule
  metadata:
    name: function
    namespace: $NAMESPACE
  spec:
    gateway: namespace-name/httpbin-gateway #The value corresponds to the Gateway CR you created. 
    service:
      name: function
      port: 80
      host: function-example.$DOMAIN
    rules:
      - path: /function
        methods: ["GET"]
        accessStrategies:
          - handler: oauth2_introspection
            config:
              required_scope: ["read"]
  EOF
  ```

>**NOTE:** If you are running Kyma on k3d, add `httpbin.kyma.local` to the entry with k3d IP in your system's `/etc/hosts` file.

The exposed Function requires all `GET` requests to have a valid token with the "read" scope.

  </details>
</div>

>**CAUTION:** When you secure a service, don't create overlapping Access Rules for paths. Doing so can cause unexpected behavior and reduce the security of your implementation.

## Access the secured resources

Follow the instructions in the tabs to call the secured service or Functions using the tokens issued for the client you registered.

<div tabs>

  <details>
  <summary>
  Call secured endpoints of a service
  </summary>

1. Send a `GET` request with a token that has the "read" scope to the HttpBin service:

  ```shell
  curl -ik -X GET https://httpbin.$DOMAIN/headers -H "Authorization: Bearer $ACCESS_TOKEN_READ"
  ```

2. Send a `POST` request with a token that has the "write" scope to the HttpBin's `/post` endpoint:

  ```shell
  curl -ik -X POST https://httpbin.$DOMAIN/post -d "test data" -H "Authorization: bearer $ACCESS_TOKEN_WRITE"
  ```

These calls return the code `200` response. If you call the service without a token, you get the code `401` response. If you call the service or its secured endpoint with a token with the wrong scope, you get the code `403` response.

  </details>

  <details>
  <summary>
  Call the secured Function
  </summary>

Send a `GET` request with a token that has the "read" scope to the Function:

  ```shell
  curl -ik https://function-example.$DOMAIN/function -H "Authorization: bearer $ACCESS_TOKEN_READ"
  ```

This call returns the code `200` response. If you call the Function without a token, you get the code `401` response. If you call the Function with a token with the wrong scope, you get the code `403` response.

  </details>
</div>

> **TIP:** To learn more about the security options read the document describing [authorization configuration](../../05-technical-reference/apix-01-config-authorizations-apigateway.md).