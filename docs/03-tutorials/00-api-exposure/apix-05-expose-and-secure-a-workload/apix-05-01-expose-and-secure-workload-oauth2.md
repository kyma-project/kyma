---
title: Expose and secure a workload with OAuth2
---

This tutorial shows how to expose and secure services or Functions using API Gateway Controller. The controller reacts to an instance of the APIRule custom resource (CR) and creates an Istio VirtualService and [Oathkeeper Access Rules](https://www.ory.sh/docs/oathkeeper/api-access-rules) according to the details specified in the CR. To interact with the secured services, the tutorial uses an OAuth2 client registered through the Hydra Maester controller.

## Prerequisites

* Deploy [a sample HttpBin service and a sample Function](./apix-01-create-workload.md).
* Set up [your custom domain](./apix-02-setup-custom-domain-for-workload.md) or use a Kyma domain instead. 
* Depending on whether you use your custom domain or a Kyma domain, export the necessary values as environment variables:
  
  <div tabs name="export-values">

    <details>
    <summary>
    Custom domain
    </summary>
    
    ```bash
    export GATEWAY=$NAMESPACE/httpbin-gateway
    ```
    </details>

    <details>
    <summary>
    Kyma domain
    </summary>

    ```bash
    export DOMAIN_TO_EXPOSE_WORKLOADS={KYMA_DOMAIN_NAME}
    export GATEWAY=kyma-system/kyma-gateway
    ```
    </details>
  </div>  

## Register an OAuth2 client and get tokens

1. Export the client name as an environment variable:

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

3. Export the client's credentials as environment variables. Run:

   ```shell
   export CLIENT_ID="$(kubectl get secret -n $NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_id}' | base64 --decode)"
   export CLIENT_SECRET="$(kubectl get secret -n $NAMESPACE $CLIENT_NAME -o jsonpath='{.data.client_secret}' | base64 --decode)"
   ```

4. Encode the client's credentials and export them as environment variables:

   ```shell
   export ENCODED_CREDENTIALS=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
   ```

5. Get tokens to interact with secured resources using the client credentials flow:

   <div tabs>
     <details>
     <summary>
     Token with `read` scope
     </summary>

     * Export the following value as an environment variable:

        ```shell
        export KYMA_DOMAIN={KYMA_DOMAIN_NAME}
        ```  

     * Get the token:

         ```shell
         curl -ik -X POST "https://oauth2.$KYMA_DOMAIN/oauth2/token" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=read"
         ```

     * Export the issued token as an environment variable:

         ```shell
         export ACCESS_TOKEN_READ={ISSUED_READ_TOKEN}
         ```

     </details>
     <details>
     <summary>
     Token with `write` scope
     </summary>

     * Export the following value as an environment variable:

        ```shell
        export KYMA_DOMAIN={KYMA_DOMAIN_NAME}
        ```  

     * Get the token:

         ```shell
         curl -ik -X POST "https://oauth2.$KYMA_DOMAIN/oauth2/token" -H "Authorization: Basic $ENCODED_CREDENTIALS" -F "grant_type=client_credentials" -F "scope=write"
         ```

     * Export the issued token as an environment variable:

         ```shell
         export ACCESS_TOKEN_WRITE={ISSUED_WRITE_TOKEN}
         ```

      </details>
   </div>

## Expose and secure your workload

Follow the instructions to expose an instance of the HttpBin service or a sample Function, and secure them with Oauth2 scopes.

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Expose the service and secure it by creating an APIRule CR in your Namespace. Run:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: httpbin
     namespace: $NAMESPACE
   spec:
     gateway: $GATEWAY
     host: httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS
     service:
       name: httpbin
       port: 8000
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

  The exposed service requires tokens with `read` scope for `GET` requests in the entire service, and tokens with `write` scope for `POST` requests to the `/post` endpoint of the service.

  </details>

  <details>
  <summary>
  Function
  </summary>

1. Expose the Function and secure it by creating an APIRule CR in your Namespace. Run:

   ```shell
   cat <<EOF | kubectl apply -f -
   apiVersion: gateway.kyma-project.io/v1beta1
   kind: APIRule
   metadata:
     name: function
     namespace: $NAMESPACE
   spec:
     gateway: $GATEWAY
     host: function-example.$DOMAIN_TO_EXPOSE_WORKLOADS
     service:
       name: function
       port: 80
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

   The exposed Function requires all `GET` requests to have a valid token with the `read` scope.

  </details>
</div>

>**CAUTION:** When you secure a workload, don't create overlapping Access Rules for paths. Doing so can cause unexpected behavior and reduce the security of your implementation.

## Access the secured resources

Follow the instructions to call the secured service or Functions using the tokens issued for the client you registered.

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

1. Send a `GET` request with a token that has the `read` scope to the HttpBin service:

   ```shell
   curl -ik -X GET https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/headers -H "Authorization: Bearer $ACCESS_TOKEN_READ"
   ```

2. Send a `POST` request with a token that has the `write` scope to the HttpBin's `/post` endpoint:

   ```shell
   curl -ik -X POST https://httpbin.$DOMAIN_TO_EXPOSE_WORKLOADS/post -d "test data" -H "Authorization: bearer $ACCESS_TOKEN_WRITE"
   ```

If successful, the call returns the code `200 OK` response. If you call the service without a token, you get the code `401` response. If you call the service or its secured endpoint with a token with the wrong scope, you get the code `403` response.

  </details>

  <details>
  <summary>
  Function
  </summary>

Send a `GET` request with a token that has the `read` scope to the Function:

   ```shell
   curl -ik https://function-example.$DOMAIN_TO_EXPOSE_WORKLOADS/function -H "Authorization: bearer $ACCESS_TOKEN_READ"
   ```

If successful, the call returns the code `200 OK` response. If you call the Function without a token, you get the code `401` response. If you call the Function with a token with the wrong scope, you get the code `403` response.

  </details>
</div>

To learn more about the security options, read the document describing [authorization configuration](../../../05-technical-reference/apix-01-config-authorizations-apigateway.md).
