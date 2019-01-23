---
title: Register a secured API
type: Tutorials
---

The Application Registry allows you to register a secured API for every service. The supported authentication methods are [Basic Authentication](https://tools.ietf.org/html/rfc7617) and [OAuth](https://tools.ietf.org/html/rfc6750).

You can specify only one authentication method for every secured API you register. If you try to register and specify more than one authentication method, the Application Registry returns a `400` code response.

## Register a Basic Authentication-secured API

To register an API secured with Basic Authentication, add a `credentials.basic` object to the `api` section of the service registration request body. You must include these fields:

| Field   |  Description |
|:----------:|:------|
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
|:----------:|:------|
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
