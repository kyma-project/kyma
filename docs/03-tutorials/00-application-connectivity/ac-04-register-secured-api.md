---
title: Register a secured API
---
<!-- TODO: Adjust to the new flow -->

Application Registry allows you to register a secured API for every service. The supported authentication methods are [Basic Authentication](https://tools.ietf.org/html/rfc7617), [OAuth](https://tools.ietf.org/html/rfc6750) (Client Credentials Grant), and client certificates.

You can specify only one authentication method for every secured API you register. 

Additionally, you can secure the API against cross-site request forgery (CSRF) attacks. CSRF tokens are an additional layer of protection and can accompany any authentication method.

>**NOTE:** Registering a secured API is a part of registering services of an external solution connected to Kyma. To learn more about this process, follow the [tutorial](ac-03-register-manage-services.md).

## Register a secured API

To register a secured API, add a **service** object to the **services** section of the Application CRD. You must include these fields:

| Field   |  Description |
|----------|------|
| **id** | Identifier of the service. Must be unique in the scope of Application CRD |
| **name** | Name of the service. Must be unique in the scope of Application CRD |
| **displayName** | Display name of the service. Must be unique in the scope of Application CRD |
| **description** | Descripion of the service |
| **providerDisplayName** | Name of the service provider |
| **entries** | Object containing service details |

**Entries** object must contain the following fields:

| Field           | Description                             |
| --------------- | --------------------------------------- |
| **credentials** | Object describing authentication method |
| **targetUrl**   | URL to the API                          |
| **type**        | Entry type                              |

**Credentials** object must contain the following fields:

| Field                 | Description                                        |
| --------------------- | -------------------------------------------------- |
| **secretName**        | Name of a secret storing credentials               |
| **type**              | Authentication method                              |
| **authenticationUrl** | Optional OAuth token URL valid only for OAuth type |

## Register a  Basic Authentication-secured API

This is an example of the service section for an API secured with Basic Authentication:

```json
  services:
  - description: "My service"
    name: my-basic-auth-service
    displayName: my-basic-auth-service
    entries:
    - credentials:
        secretName: {MY_SECRET_NAME}
        type: Basic
      targetUrl: {MY_API_URL}
      type: API
    id: 721da9cc-616e-4558-b4dc-4b58554ce7ee
    providerDisplayName: "My organisation"
  skipVerify: false
```

This is an example of secret containing credentials: 

```bash
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  username: {MY_USER_NAME}
  password: {MY_PASSWORD}
```



## Register an OAuth-secured API

This is an example of the service section for an API secured with OAuth:

```json
  services:
  - description: "My service"
    name: my-basic-auth-service
    displayName: my-basic-auth-service
    entries:
    - credentials:
        secretName: {MY_SECRET_NAME}
        authenticationUrl: {MY_OAUTH_TOKEN_URL}
        type: OAuth
      targetUrl: {MY_API_URL}
      type: API
    id: 721da9cc-616e-4558-b4dc-4b58554ce7ee
    providerDisplayName: "My organisation"
  skipVerify: false
```

This is an example of secret containing credentials: 

```bash
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  clientId: {MY_CLIENT_ID}
  clientSecret: {MY_CLIENT_SECRET}
```



## Register a client certificate-secured API

To register an API and secure it with client certificates, you must add the **credentials.certificateGen** object to the **api** section of the service registration request body. Application Registry generates a ready to use certificate and key pair for every API registered this way. You can use the generated pair or replace it with your own certificate and key.

Include this field in the service registration request body:

| Field   |  Description |
|----------|------|
| **commonName** |  Name of the generated certificate. Set as the **CN** field of the certificate Subject.  |

This is an example of the `api` section of the request body for an API secured with generated client certificates:

```json
  services:
  - description: "My service"
    name: my-basic-auth-service
    displayName: my-basic-auth-service
    entries:
    - credentials:
        secretName: {MY_SECRET_NAME}
        type: CertificateGen
      targetUrl: {MY_API_URL}
      type: API
    id: 721da9cc-616e-4558-b4dc-4b58554ce7ee
    providerDisplayName: "My organisation"
  skipVerify: false
```

>**NOTE:** If you update the registered API and change the **certificateGen.commonName**, Application Registry generates a new certificate-key pair for that API. When you delete an API secured with generated client certificates, Application Registry deletes the corresponding certificate and key.

This is an example of secret containing credentials: 

```bash
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  crt: {MY_CERTIFICATE}
  key: {MY_PRIVATE_KEY}
```

## Register a CSRF-protected API

Application Registry supports CSRF tokens as an additional layer of API protection. To register a CSRF-protected API, add the **credentials.{AUTHENTICATION_METHOD}.csrfInfo** object to the **api** section of the service registration request body.

Include this field in the service registration request body:

| Field | Description |
|-----|-----------|
| **tokenEndpointURL** | The URL to the upstream service endpoint that exposes CSRF tokens. |

This is an example of the **api** section of the request body for an API secured with both Basic Authentication and a CSRF token.

```json
 services:
  - id: {MY_UNIQUE_ID} 
    name: my-basic-auth-service
    displayName: my-basic-auth-service 
    description: "My service"
    entries:
    - credentials:
        secretName: {MY_SECRET_NAME}
        type: Basic
        csrfInfo:
          tokenEndpointURL: {www.example.com}
      targetUrl: {MY_API_URL}
      type: API
    providerDisplayName: "My organisation"
  skipVerify: false
```


## Use headers and query parameters for custom authentication

You can specify additional headers and query parameters to inject to requests made to the target API.

This is an example of the **api** section of the request body for an API secured with Basic Authentication.

```json
  services:
  - description: "My service"
    name: my-basic-auth-service
    displayName: my-basic-auth-service
    entries:
    - credentials:
        secretName: {MY_SECRET_NAME}
        type: Basic
      targetUrl: {MY_API_URL}
      requestParametersSecretName: {MY_REQ_PARAMS_SECRET_NAME}
      type: API
    id: 721da9cc-616e-4558-b4dc-4b58554ce7ee
    providerDisplayName: "My organisation"
  skipVerify: false
```

This is an example of secret containing credentials: 

```bash
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  username: {MY_USER_NAME}
  password: {MY_PASSWORD}
```

This is an example of secret containing headers and request parameters: 

```bash
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  headers: {MY_HEADERS}
  queryParameters: {MY_QUERY_PARAMETERS}
```

