---
title: Security
type: Details
---

When you expose a service in Kyma, you can secure it by specifying the **authentication** attribute in the custom resource (CR). To successfully secure the exposed service, its Pods must have the Istio sidecar injection enabled. Additionally, you must specify all of these attributes in the CR:
  - **authentication.type**
  - **jwt.issuer**
  - **jwt.jwksUri**

If you don't specify any of these attributes, the API Controller doesn't create an Istio Authentication Policy for the service and leaves it unsecured.


## Call a secured service

You can secure the exposed service using JWT authentication. This means that you must include a valid JWT ID token in the `Authorization` header of the request when you call
a secured service.

This is an example of a call to a secure exposed service:
```
curl -i https://httpbin.org/headers -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjFmNThlNTBhODI4OWMzYWM5MmE5ZTA2ZmM0YzIyZDc1NTU4MTc5YjIifQ.eyJpc3MiOiJodHRwczovL2RleC55ZmFjdG9yeS5zYXAuY29ycCIsInN1YiI6IkNpUXhPR0U0TmpnMFlpMWtZamc0TFRSaU56TXRPVEJoT1MwelkyUXhOall4WmpVME5qTVNCV3h2WTJGcyIsImF1ZCI6WyJreW1hLWNsaWVudCIsImt1YmVjb250cm9sbGVyIl0sImV4cCI6MTUzMDA5ODg3MiwiaWF0IjoxNTMwMDEyNDcyLCJhenAiOiJrdWJlY29udHJvbGxlciIsImF0X2hhc2giOiJ5QzJwY0ZmVWYzWVd2N2U5QUY3U0t3IiwiZW1haWwiOiJhZG1pbkBreW1hLmN4IiwiZW1haWxfdmVyaWZpZWQiOnRydWUsIm5hbWUiOiJhZG1pbiJ9.pxy4P95PVSwIiXArcfsqAPVFhBmo5sHzUnqzwY6HF9UgMRkDFlIs5CKe1ZiGteGr6-gYU_0VmHroZ4alpcVcpL8Z5M2xnlOaZDyB8TNLvUAATpElcBMy6Cxb_7zLwP91IsX0QgI3DTg3H-M0eaJ4VwMKrfEu9h2rwxzvBrDc5vB_1Bm8OABl08wLSQpR27GGsI58RmA5YJmZX1PSzv90Zl_krqyvWIe6pmcHCrP--02LLUaoxhY42IDWkF8n9RPMLixmFZFXbeonCddR30OUkAbEFLBVBf8nJaFDms_VjZHSXZDitCu4r6myE4AnT_IeXI2dRgdGT73Hh8895zu7fQ"
```

## Specify multiple JWT token issuers

You can specify multiple JWT token issuers to allow to access the secured service with tokens issued by different ID providers. You can successfully call the secured service using JWT ID tokens issued by any of the parties specified in the **authentication** attribute of the CR. This is an example of the **authentication** attribute that allows to access the service using JWT tokens signed by two different issuers.

```
- type: JWT
  jwt:
    issuer: https://sampleissuer1.abc.com
    jwksUri: https://www.sampleapis.com/oauth2/v3/certs
- type: JWT
  jwt:
    issuer: https://sampleissuer2.abc.com
    jwksUri: https://www.regularsampleapis.com/oauth2/v3/certs
```

## Specify service resource paths not secured with JWT authentication

>**CAUTION:** Resource paths excluded from JWT validation are accessible publicly and anonymously without a token.

To disable JWT authentication for specific resource paths of a given service, add a list of path-matching expressions to the `triggerRule.excludedPaths` attribute in the [Api CR](#custom-resource-api) of the service.   
The syntax for path-matching expressions is `{EXPRESSION_TYPE}: {VALUE}`. The `{VALUE}` is a string and there are four available expression types:

- `exact`: must match the specified path exactly.
- `prefix`: must match the path prefix.
- `suffix`: must match the path suffix.
- `regex`: must match the path with a regular expression as defined in [this](https://en.cppreference.com/w/cpp/regex/ecmascript) specification.

This is an example path-matching expressions list that defines resource paths that are not secured with JWT authentication:

```
    triggerRule:
      excludedPaths:
      - exact: /do/not/use/in/production
      - exact: /status
      - suffix: index.html
      - suffix: /favicon.ico
      - regex: ^/web/static/(.*?)?
      - regex: ^/api/users/[a-zA-Z0-9]+/avatar.png
```
