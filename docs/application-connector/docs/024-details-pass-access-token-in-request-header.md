---
title: Pass an access token in a request header
type: Details
---

The Application Connector supports passing the access token directly in the request.

## Passing the access token

If the user is already authenticated to the service deployed on Kyma, the access token can be passed in a custom `Access-token` header. If the Application Connector detects that the custom header is present, instead of obtaining a new token, it passes the received one as a `Bearer` token in the `Authorization` header.

## Example

Find the example of passing the EC access token to the Application Connector using lambda in the [`examples`](https://github.com/kyma-project/examples/tree/master/call-ec) repository.
