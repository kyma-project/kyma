# Register a Secured API

Application Connectivity allows you to register a secured API exposed by your external solution. The supported authentication methods are [Basic Authentication](https://tools.ietf.org/html/rfc7617), [OAuth](https://tools.ietf.org/html/rfc6750), [OAuth 2.0 mTLS](https://datatracker.ietf.org/doc/html/rfc8705), and client certificates.

You can specify only one authentication method for every secured API you register.

Additionally, you can secure the API against cross-site request forgery (CSRF) attacks. CSRF tokens are an additional layer of protection and can accompany any authentication method.

> [!NOTE]
> Registering a secured API is a part of [registering services](01-20-register-manage-services.md) of an external solution connected to Kyma.

## Register a Secured API

To register a secured API, add a **service** object to the **services** section of the Application CR. You must include these fields:

| Field   |  Description |
|----------|------|
| **id** | Identifier of the service. Must be unique in the scope of Application CR. |
| **name** | Name of the service. Must be unique in the scope of Application CR. Allowed characters include: lowercase letters, numbers, and hyphens. |
| **displayName** | Display name of the service. Must be unique in the scope of Application CR. Its normalized form constitutes a part of the **GATEWAY_URL** path. The normalized version of **displayName** in the path is stripped of all non-lowercase, non-alphanumeric characters except hyphens, and of all trailing hyphens. |
| **description** | Description of the service |
| **providerDisplayName** | Name of the service provider |
| **entries** | Object containing service details |

### Entries

The **entries** object must contain the following fields:

| Field                           | Description                                                  |
| ------------------------------- | ------------------------------------------------------------ |
| **credentials**                 | Optional object containing credentials used for authentication. Must be specified for secured APIs. |
| **targetUrl**                   | URL to the API.                                               |
| **type**                        | Entry type. Use the `API` type when registering an API.        |
| **requestParametersSecretName** | Optional name of a Secret with additional request parameters and headers. |

#### Credentials

The **credentials** object must contain the following fields:

| Field                 | Description                                                                 |
| --------------------- |-----------------------------------------------------------------------------|
| **secretName**        | Name of a Secret storing credentials.                                        |
| **type**              | Authentication method type. Supported values: `Basic`, `OAuth`, `OAuthWithCert`, `CertificateGen`.  |
| **authenticationUrl** | Optional OAuth token URL, valid only for the `OAuth` and `OAuthWithCert` types. |

## Register a Basic Authentication-secured API

This is an example of the **service** object for an API secured with Basic Authentication:

   ```yaml
     - id: {TARGET_UUID}
       name: my-basic-auth-service
       displayName: "My Basic Auth Service"
       description: "My service"
       providerDisplayName: "My organisation"
       entries:
       - credentials:
           secretName: {SECRET_NAME}
           type: Basic
         targetUrl: {TARGET_API_URL}
         type: API
   ```

This is an example Secret containing credentials:

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: {SECRET_NAME}
     namespace: kyma-system
   data:
     username: {BASE64_ENCODED_USER_NAME}
     password: {BASE64_ENCODED_PASSWORD}
   ```

To create such a Secret, run this command:

   ```bash
   kubectl create secret generic {SECRET_NAME} --from-literal username={USER_NAME} --from-literal password={PASSWORD} -n kyma-system
   ```

## Register an OAuth-secured API

This is an example of the **service** object for an API secured with OAuth:

   ```yaml
     - id: {TARGET_UUID}
       name: my-oauth-service
       displayName: "My OAuth Service"    
       description: "My service"
       providerDisplayName: "My organisation"
       entries:
       - credentials:
           secretName: {SECRET_NAME}
           authenticationUrl: {OAUTH_TOKEN_URL}
           type: OAuth
         targetUrl: {TARGET_API_URL}
         type: API
   ```

This is an example of the Secret containing credentials:

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: {SECRET_NAME}
     namespace: kyma-system
   data:
     clientId: {BASE64_ENCODED_CLIENT_ID}
     clientSecret: {BASE64_ENCODED_CLIENT_SECRET}
   ```

To create such a Secret, run this command:

   ```bash
   kubectl create secret generic {SECRET_NAME} --from-literal clientId={CLIENT_ID} --from-literal clientSecret={CLIENT_SECRET} -n kyma-system
   ```

## Register an OAuth 2.0 mTLS-secured API

This is an example of the **service** object for an API secured with OAuth where the token is fetched from an mTLS-secured endpoint:

   ```yaml
     - id: {TARGET_UUID}
       name: my-mTLS-oauth-service
       displayName: "My mTLS OAuth Service"    
       description: "My service"
       providerDisplayName: "My organisation"
       entries:
       - credentials:
           secretName: {SECRET_NAME}
           authenticationUrl: {OAUTH_TOKEN_URL}
           type: OAuthWithCert
         targetUrl: {TARGET_API_URL}
         type: API
   ```

This is an example of the Secret containing credentials:

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: {SECRET_NAME}
     namespace: kyma-system
   data:
     clientId: {BASE64_ENCODED_CLIENT_ID}
     crt: {BASE64_ENCODED_CERTIFICATE}
     key: {BASE64_ENCODED_PRIVATE_KEY}
   ```

To create such a Secret, run this command:

   ```bash
   kubectl create secret generic {SECRET_NAME} --from-literal clientId={CLIENT_ID} --from-literal crt={CERTIFICATE} --from-literal key={PRIVATE_KEY} -n kyma-system
   ```

## Register a Client Certificate-Secured API

This is an example of the **service** object for an API secured with a client certificate:

   ```yaml
     - id: {TARGET_UUID}
       name: my-client-cert-service
       displayName: "My Client Cert Service"
       description: "My service"
       providerDisplayName: "My organisation"
       entries:
       - credentials:
           secretName: {SECRET_NAME}
           type: CertificateGen
         targetUrl: {TARGET_API_URL}
         type: API
   ```

This is an example of the Secret containing credentials:

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: {SECRET_NAME}
     namespace: kyma-system
   data:
     crt: {BASE64_ENCODED_CERTIFICATE}
     key: {BASE64_ENCODED_PRIVATE_KEY}
   ```

To create such a Secret, run this command:

   ```bash
   kubectl create secret generic {SECRET_NAME} --from-literal crt={CERTIFICATE} --from-literal key={PRIVATE_KEY} -n kyma-system
   ```

## Register a CSRF-protected API

This is an example of the **service** object for an API secured with both Basic Authentication and a CSRF token:

   ```yaml
     - id: {TARGET_UUID} 
       name: my-csrf-service
       displayName: "My CSRF Service" 
       description: "My service"
       providerDisplayName: "My organisation"
       entries:
       - credentials:
           secretName: {SECRET_NAME}
           type: Basic
           csrfInfo:
             tokenEndpointURL: {CSRF_TOKEN_URL}
         targetUrl: {TARGET_API_URL}
         type: API
   ```

> [!NOTE]
> The example assumes that the CSRF token endpoint service uses the same credentials as the target API.

This is an example of the Secret containing credentials:

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: {SECRET_NAME}
     namespace: kyma-system
   data:
     username: {BASE64_ENCODED_USER_NAME}
     password: {BASE64_ENCODED_PASSWORD}
   ```

To create such a Secret, run this command:

   ```bash
   kubectl create secret generic {SECRET_NAME} --from-literal username={USER_NAME} --from-literal password={PASSWORD} -n kyma-system
   ```

## Use Headers and Query Parameters for Custom Authentication

You can specify additional headers and query parameters to inject to requests made to the target API. You can use it with any authentication method.

This is an example of the **service** object for an API secured with Basic Authentication and including additional headers and query parameters.

   ```yaml
     - id: {TARGET_UUID}
       name: my-headers-params-service
       displayName: "My Headers Params Service"
       description: "My service"
       providerDisplayName: "My organisation"
       entries:
       - credentials:
           secretName: {SECRET_NAME}
           type: Basic
         targetUrl: {TARGET_API_URL}
         requestParametersSecretName: {QUERY_PARAMS_SECRET_NAME}
         type: API
   ```

This is an example of the Secret containing credentials:

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: {SECRET_NAME}
     namespace: kyma-system
   data:
     username: {BASE64_ENCODED_USER_NAME}
     password: {BASE64_ENCODED_PASSWORD}
   ```

To create such a Secret, run this command:

   ```bash
   kubectl create secret generic {SECRET_NAME} --from-literal username={USER_NAME} --from-literal password={PASSWORD} -n kyma-system
   ```

This is an example of the Secret containing headers and request parameters:

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: {SECRET_NAME}
     namespace: kyma-system
   data:
     headers: {BASE64_ENCODED_HEADERS_JSON}
     queryParameters: {BASE64_ENCODED_QUERY_PARAMS_JSON}
   ```

To create such a Secret, run this command:

   ```bash
   kubectl create secret generic {SECRET_NAME} --from-literal headers={HEADERS_JSON} --from-literal queryParameters={QUERY_PARAMS_JSON} -n kyma-system
   ```

Additional headers stored in the Secret must be provided in the form of a valid JSON document. This is an example of a headers JSON containing one entry:

   ```json
   {"{MY_HEADER}":["{MY_HEADER_VALUE}"]}
   ```

Additional query parameters stored in the Secret must be provided in the form of a valid JSON document. This is an example of a headers JSON containing one entry:

   ```json
   {"{MY_QUERY_PARAM}":["{MY_QUERY_PARAM_VALUE}"]}
   ```
