---
title: Architecture
---

The following diagram illustrates the authorization and authentication flow in Kyma. The representation assumes the Kyma Console UI as the user's point of entry.

![authorization-authentication-flow](./assets/001-kyma-authorization.svg)

1. The user opens the Kyma Console UI. If the Console application doesn't find a JWT token in the browser session storage, it redirects the user's browser to the Open ID Connect (OIDC) provider, Dex.
2. Dex lists all defined Identity Provider connectors to the user. The user selects the Identity Provider to authenticate with. After successful authentication, the browser is redirected back to the OIDC provider which issues a JWT token to the user. After obtaining the token, the browser is redirected back to the Console UI. The Console UI stores the token in the Session Storage and uses it for all subsequent requests.
3. The Authorization Proxy validates the JWT token passed in the **Authorization Bearer** request header. It extracts the user and groups details, the requested resource path, and the request method from the token. The Proxy uses this data to build an attributes record, which it sends to the Kubernetes Authorization API.
4. The Proxy sends the attributes record to the Kubernetes Authorization API. If the authorization fails, the flow ends with a `403` code response.
5. If the authorization succeeds, the request is forwarded to the Kubernetes API Server.  

>**NOTE:** The Authorization Proxy can verify JWT tokens issued by Dex because Dex is registered as a trusted issuer through OIDC parameters during the Kyma installation.  
