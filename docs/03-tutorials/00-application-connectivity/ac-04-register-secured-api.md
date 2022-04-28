---
title: Register a secured API
---
You can register a secured API exposed by your external soliution. The supported authentication methods are [Basic Authentication](https://tools.ietf.org/html/rfc7617), [OAuth](https://tools.ietf.org/html/rfc6750), and client certificates.

You can specify only one authentication method for every secured API you register. 

Additionally, you can secure the API against cross-site request forgery (CSRF) attacks. CSRF tokens are an additional layer of protection and can accompany any authentication method.

>**NOTE:** Registering a secured API is a part of registering services of an external solution connected to Kyma. To learn more about this process, follow the [tutorial](ac-03-register-manage-services.md).

## Register a secured API

To register a secured API, add a **service** object to the **services** section of the Application CRD. You must include these fields:

| Field   |  Description |
|----------|------|
| **id** | Identifier of the service. Must be unique in the scope of Application CRD |
| **name** | Name of the service. Must be unique in the scope of Application CRD. Allowed characters include: lowercase letters, numbers, and hyphens. |
| **displayName** | Display name of the service. Must be unique in the scope of Application CRD. Its normalized form constitutes a part of the path. |
| **description** | Descripion of the service |
| **providerDisplayName** | Name of the service provider |
| **entries** | Object containing service details |

**Entries** object must contain the following fields:

| Field                           | Description                                                  |
| ------------------------------- | ------------------------------------------------------------ |
| **credentials**                 | Optional object containing credentials used for authentication. Must be specified for secured APIs. |
| **targetUrl**                   | URL to the API                                               |
| **type**                        | Entry type. Use **API** type when registering an API.        |
| **requestParametersSecretName** | Optional name of a secret with additional request parameters and headers |

**Credentials** object must contain the following fields:

| Field                 | Description                                                  |
| --------------------- | ------------------------------------------------------------ |
| **secretName**        | Name of a secret storing credentials                         |
| **type**              | Authentication method type. Supported values: **Basic**, **OAuth**, **CertGen**. |
| **authenticationUrl** | Optional OAuth token URL valid only for OAuth type           |

## Register a  Basic Authentication-secured API

This is an example of the **service** object for an API secured with Basic Authentication:

```yaml
  - id: {MY_UNIQUE_ID}
    name: my-basic-auth-service
    displayName: "My Basic Auth Service"
    description: "My service"
    providerDisplayName: "My organisation"
    entries:
    - credentials:
        secretName: {MY_SECRET_NAME}
        type: Basic
      targetUrl: {MY_API_URL}
      type: API
```

This is an example of a secret containing credentials: 

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  username: {MY_USER_NAME_BASE64_ENCODED}
  password: {MY_PASSWORD_BASE64_ENCODED}
```

You can execute this command to create the secret:

```bash
kubectl create secret generic {MY_SECRET_NAME} --from-literal username={MY_USER_NAME} --from-literal password={MY_PASSWORD} -n kyma-integration
```

## Register an OAuth-secured API

This is an example of the **service** object for an API secured with OAuth:

```yaml
  - id: {MY_UNIQUE_ID}
    name: my-oauth-service
    displayName: "My OAuth Service"    
    description: "My service"
    providerDisplayName: "My organisation"
    entries:
    - credentials:
        secretName: {MY_SECRET_NAME}
        authenticationUrl: {MY_OAUTH_TOKEN_URL}
        type: OAuth
      targetUrl: {MY_API_URL}
      type: API
```

This is an example of a secret containing credentials: 

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  clientId: {MY_CLIENT_ID_BASE64_ENCODED}
  clientSecret: {MY_CLIENT_SECRET_BASE64_ENCODED}
```

You can execute this command to create the secret:

```bash
kubectl create secret generic {MY_SECRET_NAME} --from-literal clientId={MY_CLIENT_ID} --from-literal clientSecret={MY_CLIENT_SECRET} -n kyma-integration
```

## Register a client certificate-secured API

This is an example of the **service** object for an API secured with certificate:

```yaml
  - id: {MY_UNIQUE_ID}
    name: my-client-cert-service
    displayName: "My Client Cert Service"
    description: "My service"
    providerDisplayName: "My organisation"
    entries:
    - credentials:
        secretName: {MY_SECRET_NAME}
        type: CertificateGen
      targetUrl: {MY_API_URL}
      type: API
```

This is an example of a secret containing credentials: 

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  crt: {MY_CERTIFICATE_BASE64_ENCODED}
  key: {MY_PRIVATE_KEY_BASE64_ENCODED}
```

You can execute this command to create the secret:

```bash
kubectl create secret generic {MY_SECRET_NAME} --from-literal crt={MY_CERTIFICATE} --from-literal key={MY_PRIVATE_KEY} -n kyma-integration
```

## Register a CSRF-protected API

This is an example of the **service** object for an API secured with both Basic Authentication and a CSRF token:

```yaml
  - id: {MY_UNIQUE_ID} 
    name: my-csrf-service
    displayName: "My CSRF Service" 
    description: "My service"
    providerDisplayName: "My organisation"
    entries:
    - credentials:
        secretName: {MY_SECRET_NAME}
        type: Basic
        csrfInfo:
          tokenEndpointURL: {MY_CSRF_TOKEN_URL}
      targetUrl: {MY_API_URL}
      type: API
```

> **NOTE:** it is assumed that CSRF token endpoint service uses the same credentials as the target API.

This is an example of a secret containing credentials: 

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  username: {MY_USER_NAME_BASE64_ENCODED}
  password: {MY_PASSWORD_BASE64_ENCODED}
```

You can execute this command to create the secret:

```bash
kubectl create secret generic {MY_SECRET_NAME} --from-literal username={MY_USER_NAME} --from-literal password={MY_PASSWORD} -n kyma-integration
```

## Use headers and query parameters for custom authentication

You can specify additional headers and query parameters to inject to requests made to the target API. You can use it with any authentication method.

This is an example of the **service** object for an API secured with Basic Authentication and including additional headers and query parameters.

```yaml
  - id: {MY_UNIQUE_ID}
    name: my-headers-params-service
    displayName: "My Headers Params Service"
    description: "My service"
    providerDisplayName: "My organisation"
    entries:
    - credentials:
        secretName: {MY_SECRET_NAME}
        type: Basic
      targetUrl: {MY_API_URL}
      requestParametersSecretName: {MY_REQ_PARAMS_SECRET_NAME}
      type: API
```

This is an example of a secret containing credentials: 

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  username: {MY_USER_NAME_BASE64_ENCODED}
  password: {MY_PASSWORD_BASE64_ENCODED}
```

You can execute this command to create the secret:

```bash
kubectl create secret generic {MY_SECRET_NAME} --from-literal username={MY_USER_NAME} --from-literal password={MY_PASSWORD} -n kyma-integration
```

This is an example of a secret containing headers and request parameters: 

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {MY_SECRET_NAME}
  namespace: kyma-integration
data:
  headers: {MY_HEADERS_JSON_BASE64_ENCODED}
  queryParameters: {MY_QUERY_PARAMETERS_JSON_BASE64_ENCODED}
```

 You can execute this command to create the secret:

```bash
kubectl create secret generic {MY_SECRET_NAME} --from-literal headers={MY_HEADERS} --from-literal headers={MY_QUERY_PARAMETERS} -n kyma-integration
```

Additional headers stored in the secret must be provided in a form of a valid JSON document. This is an example of headers JSON containing one entry:

```json
{"{MY_HEADER}":["{MY_VALUE}"]}
```

Additional request parameters stored in the secret must be provided in a form of a valid JSON document. This is an example of headers JSON containing one entry:

```json
{"{MY_REQUEST_PARAM}":["{MY_REQUEST_PARAM}"]}
```

 
