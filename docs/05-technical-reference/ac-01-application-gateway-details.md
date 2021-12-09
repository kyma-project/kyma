---
title: Application Gateway details
---

Application Gateway is an intermediary component between a Function or a service and an external API.

## Proxying requests
<!-- TODO: describe the structure of the Secret storing credentials -->
Application Gateway proxies requests from Functions and services in Kyma to external APIs based on the configuration stored in Secrets.

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
