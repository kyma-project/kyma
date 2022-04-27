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

| Field          | Description                          |
| -------------- | ------------------------------------ |
| **secretName** | Name of a secret storing credentials |
| **type**       | Authentication method                |

## Register a  Basic Authentication-secured API

This is an example of the service section for an API secured with Basic Authentication:

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
      targetUrl: {MY_API_URL}
      type: API
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
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "credentials": {
            "certificateGen": {
                "commonName": "{CERT_NAME}"
            }
        }
    }
```

>**NOTE:** If you update the registered API and change the **certificateGen.commonName**, Application Registry generates a new certificate-key pair for that API. When you delete an API secured with generated client certificates, Application Registry deletes the corresponding certificate and key.

### Details

When you register an API with the **credentials.certificateGen** object, Application Registry generates a SHA256withRSA-encrypted certificate and a matching key. To enable communication between Kyma and an API secured with this authentication method, set the certificate as a valid authentication medium for all calls coming from Kyma in your external solution.

You can retrieve the client certificate by sending the following request:

```bash
curl https://gateway.$CLUSTER_DOMAIN/$APP_NAME/v1/metadata/services/$SERVICE_ID --cert $CLIENT_CERT_FILE_NAME.crt --key $KEY_FILE_NAME.key
```

> **CAUTION:** On a local Kyma deployment, skip SSL certificate verification when making a `curl` call, by adding the `-k` flag to it. Alternatively, add the Kyma certificates to your local certificate storage on your machine using the `kyma import certs` command.

A successful call returns a response body with the details of a registered service and a base64-encoded client certificate.

The certificate and key pair is stored in a Secret in the `kyma-integration` Namespace. List all Secrets and find the one created for your API:

```bash
kubectl -n kyma-integration get secrets
```

To fetch the certificate and key encoded with base64, run this command:

```bash
kubectl -n kyma-integration get secrets $APP_NAME-$SERVICE_ID -o yaml
```

>**NOTE:** `APP_NAME` is the name of the Application used to connect the external solution that is the origin of the API. `SERVICE_ID` is the ID of the registered service to which the API belongs. You get this ID after you register an external solution's service in Kyma.


If the API you registered provides a certificate-key pair or the generated certificate doesn't meet your security standards or specific needs, you can use a custom certificate-key pair for authentication. To replace the Kyma-generated pair with your certificate and key, run this command:

```bash
kubectl -n kyma-integration patch secrets $APP_NAME-$SERVICE_ID --patch 'data:
  crt: {BASE64_ENCODED_CRT}
  key: {BASE64_ENCODED_KEY}'
```

## Register a CSRF-protected API

Application Registry supports CSRF tokens as an additional layer of API protection. To register a CSRF-protected API, add the **credentials.{AUTHENTICATION_METHOD}.csrfInfo** object to the **api** section of the service registration request body.

Include this field in the service registration request body:

| Field | Description |
|-----|-----------|
| **tokenEndpointURL** | The URL to the upstream service endpoint that exposes CSRF tokens. |

This is an example of the **api** section of the request body for an API secured with both Basic Authentication and a CSRF token.

```json
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "credentials": {
            "basic": {
                "username": "{USERNAME}",
                "password": "{PASSWORD}",
                "csrfInfo": {
                    "tokenEndpointURL": "{TOKEN_ENDPOINT_URL}"
                }
            }
        }
    }
```


## Use headers and query parameters for custom authentication

You can specify additional headers and query parameters to inject to requests made to the target API.

This is an example of the **api** section of the request body for an API secured with Basic Authentication.

```json
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "requestParameters": {
            "headers": {
                "{CUSTOM_HEADER_NAME}" : ["{CUSTOM_HEADER_VALUE}"]
            },
            "queryParameters": {
                "{CUSTOM_QUERY_PARAMETER_NAME}" : ["{CUSTOM_QUERY_PARAMETER_VALUE}"]
            }
        },
        "credentials": {
            "basic": {
                "username": "{USERNAME}",
                "password": "{PASSWORD}"
            }
        }
    }
```
