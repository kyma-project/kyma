---
title: Proxying requests by the Application Gateway
type: Details
---

The Application Gateway proxies requests from lambda functions and services in Kyma to external APIs. When the Application Gateway works in the alternative `gatewayOncePerNamespace` [mode](#architecture-application-connector-components-application-operator), it proxies the requests based on the configuration stored in Secrets.
​
### Proxy configuration
​
The Secret expected by the Application Gateway has the following structure:

```
data:
    {API_NAME_1}_TARGET_URL: {BASE64_ENCODED_API1_URL}
    {API_NAME_2}_TARGET_URL: {BASE64_ENCODED_API2_URL}
    CONFIGURATION: {BASE64_ENCODED_CONFIG_JSON}
    CREDENTIALS_TYPE: {BASE64_ENCODED_CREDENTIALS_TYPE}
```

* The `BASE64_ENCODED_CONFIG_JSON` configuration contains credentials and request parameters. 
* The `BASE64_ENCODED_CREDENTIALS_TYPE` assumes one of the following values:  `OAuth`, `BasicAuth`, `Certificate` (not supported in the Director), `NoAuth`.
​
An example `CONFIGURATION` for APIs secured with OAuth looks as follows:

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

> **NOTE:** All APIs defined in a single Secret use the same configuration - the same credentials, CSRF tokens, and request parameters.
​
​
### Calling Application Gateway
​
The Secret name and the API name are specified as path variables in the following format:

```
http://my-namespace-application-gateway:8080/secret/{SECRET_NAME}/api/{API_NAME}/{CUSTOM_PATH}
```
​
In order for the Application Gateway to properly read Secrets, they must exist in the same Namespace as the Gateway does.