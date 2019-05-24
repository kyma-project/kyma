---
title: Register a secured API
type: Tutorials
---

The Application Registry allows you to register a secured API for every service. The supported authentication methods are [Basic Authentication](https://tools.ietf.org/html/rfc7617), [OAuth](https://tools.ietf.org/html/rfc6750), and client certificates.

You can specify only one authentication method for every secured API you register. If you try to register and specify more than one authentication method, the Application Registry returns a `400` code response.

Additionally, you can secure the API against cross-site request forgery (CSRF) attacks. CSRF tokens are an additional layer of protection and can accompany any authentication method.  

>**NOTE:** Registering a secured API is a part of registering services of an external solution connected to Kyma. To learn more about this process, follow [this](#tutorials-register-a-service) tutorial.

## Register a Basic Authentication-secured API

To register an API secured with Basic Authentication, add a `credentials.basic` object to the `api` section of the service registration request body. You must include these fields:

| Field   |  Description |
|----------|------|
| **username** | Basic Authorization username |
| **password** | Basic Authorization password |

This is an example of the `api` section of the request body for an API secured with Basic Authentication:

```
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "credentials": {
            "basic": {
                "username": "{USERNAME}",
                "password": "{PASSWORD}"
            },
        }  
```
## Register an OAuth-secured API

To register an API secured with OAuth, add a `credentials.oauth` object to the `api` section of the service registration request body. You must include these fields:

| Field   |  Description |
|----------|------|
| **url** |  OAuth token exchange endpoint of the service |
| **clientId** | OAuth client ID |
| **clientSecret** | OAuth client Secret |    

This is an example of the `api` section of the request body for an API secured with OAuth:

```
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "credentials": {
            "oauth": {
                "url": "https://sampleapi.targeturl/authorizationserver/oauth/token",
                "clientId": "{CLIENT_ID}",
                "clientSecret": "{CLIENT_SECRET}"
            },
        }  
```

## Register a client certificate-secured API

To register an API and secure it with client certificates, you must add the `credentials.certificateGen` object to the `api` section of the service registration request body. The Application Registry generates a ready to use certificate and key pair for every API registered this way. You can use the generated pair or replace it with your own certificate and key.

Include this field in the service registration request body:

| Field   |  Description |
|----------|------|
| **commonName** |  Name of the generated certificate. Set as the `CN` field of the certificate Subject.  |

This is an example of the `api` section of the request body for an API secured with generated client certificates:

```
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "credentials": {
            "certificateGen": {
                "commonName": "{CERT_NAME}"
            },
        }  
```

>**NOTE:** If you update the registered API and change the `certificateGen.commonName`, the Application Registry generates a new certificate-key pair for that API. When you delete an API secured with generated client certificates, the Application Registry deletes the corresponding certificate and key.

### Details

When you register an API with the `credentials.certificateGen` object, the Application Registry generates a SHA256withRSA-encrypted certificate and a matching key. To enable communication between Kyma and an API secured with this authentication method, set the certificate as a valid authentication medium for all calls coming from Kyma in your external solution.

You can retrieve the client certificate by sending the following request:
```
curl https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services/{YOUR_SERVICE_ID} --cert {CERT_FILE_NAME}.crt --key {KEY_FILE_NAME}.key -k
```
A successful call will return a response body with the details of a registered service and a base64-encoded client certificate.

The certificate and key pair is stored in a Secret in the `kyma-integration` Namespace. List all Secrets and find the one created for your API:

```
kubectl -n kyma-integration get secrets
```

To fetch the certificate and key encoded with base64, run this command:

```
kubectl -n kyma-integration get secrets app-{APP_NAME}-{SERVICE_ID} -o yaml
```

>**NOTE:** Replace the `APP_NAME` placeholder with the name of the Application used to connect the external solution that is the origin of the API. Replace the `SERVICE_ID` placeholder with the ID of the registered service to which the API belongs. You get this ID after you register an external solution's service in Kyma.


If the API you registered provides a certificate-key pair or the generated certificate doesn't meet your security standards or specific needs, you can use a custom certificate-key pair for authentication. To replace the Kyma-generated pair with your certificate and key, run this command:

```
kubectl -n kyma-integration patch secrets app-{APP_NAME}-{SERVICE_ID} --patch 'data:
  crt: {BASE64_ENCODED_CRT}
  key: {BASE64_ENCODED_KEY}'
```

## Register a CSRF-protected API

The Application Registry supports CSRF tokens as an additional layer of API protection. To register a CSRF-protected API, add the `credentials.{AUTHENTICATION_METHOD}.csrfInfo` object to the `api` section of the service registration request body.

Include this field in the service registration request body:

| Field | Description |
|-----|-----------|
| **tokenEndpointURL** | The URL to the upstream service endpoint that exposes CSRF tokens. |

This is an example of the `api` section of the request body for an API secured with both Basic Authentication and a CSRF token.

```
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "credentials": {
            "basic": {
                "username": "{USERNAME}",
                "password": "{PASSWORD}",
                "csrfInfo": {
                    "tokenEndpointURL": "{TOKEN_ENDPOINT_URL}"
                }
            },
        }
```

## Specify custom headers and query parameters for authentication requests

You can specify additional headers and query parameters that will be injected during authentication requests.

This is an example of the **api** section of the request body for an API secured with Basic Authentication. It is enriched with the **custom-header** header with the `foo` value, and the **param** query parameter with the `bar` value.

```
    "api": {
        "targetUrl": "https://sampleapi.targeturl/v1",
        "credentials": {
            "headers": {
                "custom-header": ["foo"]
            },
            "queryParameters": {
                "param": ["bar"]
            },
            "basic": {
                "username": "{USERNAME}",
                "password": "{PASSWORD}"
            },
        }
```
