---
title: Pass an access token in a request header
---

Application Connector supports passing the access token directly in the request.

## Passing the access token

If the user is already authenticated to the target API, the access token can be passed in a custom `Access-token` header. The value of the header is of the `Bearer {TOKEN}` or `Basic {TOKEN}` form. If Application Connector detects that the custom header is present, instead of performing authentication steps, it removes the `Access-token` header and passes the received value in the `Authorization` header.