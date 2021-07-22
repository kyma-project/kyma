---
title: Application Gateway details
---

>**CAUTION:** This document describes Application Gateway working in the alternative Compass [mode](03-architecture/ac-01-application-connector-components.md#application-operator). See the [architecture](03-architecture/ac-03-application-gateway.md) to learn about Application Gateway working in the default legacy mode.

Application Gateway is an intermediary component between a Function or a service and an external API.

## Proxying requests

Application Gateway proxies requests from Functions and services in Kyma to external APIs based on the configuration stored in Secrets.

### Proxy configuration

The Secret that Application Gateway expects has the following structure:

```
data:
    {API_NAME_1}_TARGET_URL: {BASE64_ENCODED_API1_URL}
    {API_NAME_2}_TARGET_URL: {BASE64_ENCODED_API2_URL}
    CONFIGURATION: {BASE64_ENCODED_CONFIG_JSON}
    CREDENTIALS_TYPE: {BASE64_ENCODED_CREDENTIALS_TYPE}
```

* The `BASE64_ENCODED_CONFIG_JSON` configuration contains credentials and request parameters.
* The `BASE64_ENCODED_CREDENTIALS_TYPE` assumes one of the following values:  `OAuth`, `BasicAuth`, `NoAuth`.

An example **CONFIGURATION** for APIs secured with OAuth looks as follows:

```json
{
    "credentials": {
        "clientId": "{OAUTH_CLIENT_ID}",
        "clientSecret": "{OAUTH_CLIENT_SECRET}",
        "requestParameters": {},
        "tokenUrl": "{OAUTH_TOKEN_URL}"
    },
    "csrfConfig": {
        "tokenUrl": "{CSRF_TOKEN_URL}"
    },
    "requestParameters": {
        "headers": {
            "Content-Type": ["application/json"]
        },
        "queryParameters": {
            "limit": ["50"]
        }
    }
}
```

An example **CONFIGURATION** for APIs secured with BasicAuth looks as follows:

```json
{
    "credentials": {
      "username":"{USERNAME}",
      "password":"{PASSWORD}"
    },
    "csrfConfig": {
        "tokenUrl": "{CSRF_TOKEN_URL}"
    },
    "requestParameters": {
        "headers": {
            "Content-Type": ["application/json"]
        },
        "queryParameters": {
            "limit": ["50"]
        }
    }
}
```

> **NOTE:** All APIs defined in a single Secret use the same configuration - the same credentials, CSRF tokens, and request parameters.


### Calling Application Gateway

The Secret name and the API name are specified as path variables in the following format:

```
http://my-namespace-application-gateway:8080/secret/{SECRET_NAME}/api/{API_NAME}/{CUSTOM_PATH}
```

In order for Application Gateway to properly read Secrets, they must exist in the same Namespace as the Gateway.

## Caching

To ensure optimal performance, Application Gateway caches the OAuth tokens and CSRF tokens it obtains. If the service doesn't find valid tokens for the call it makes, it gets new tokens from the OAuth server and the CSRF token endpoint.
Additionally, the service caches ReverseProxy objects used to proxy requests to the underlying URL.

## Handling of headers

Application Gateway removes the following headers while making calls to the registered Applications:

- `X-Forwarded-Proto`
- `X-Forwarded-For`
- `X-Forwarded-Host`
- `X-Forwarded-Client-Cert`

In addition, the `User-Agent` header is set to an empty value not specified in the call, which prevents setting the default value.