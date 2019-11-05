---
title: Security
type: Details
---

## Client certificates

To provide maximum security, the Application Connector uses TLS protocol with Client Authentication enabled. As a result, whoever wants to connect to the Application Connector must present a valid client certificate, which is dedicated to a specific Application. In this way, the traffic is fully encrypted and the client has a valid identity.

## Disable SSL certificate verification

You can disable the SSL certificate verification in the communication between Kyma and an application to allow Kyma to send requests and data to an unsecured application. Disabling the certificate verification can be useful in certain testing scenarios.

>**NOTE:** By default, the SSL certificate verification is enabled when sending data and requests to every application.

Follow these steps to disable SSL certificate verification for communication between Kyma and an external application:

  1. Edit the `{APPLICATION_CR_NAME}-application-gateway` Deployment in the `kyma-integration` Namespace. Run:
    ```
    kubectl -n kyma-integration edit deployment {APPLICATION_CR_NAME}-application-gateway
    ```
  2. Edit the Deployment in Vim. Select `i` to start editing.
  3. Find the **skipVerify** parameter and change its value to `true`.
  4. Select `esc`, type `:wq`, and select `enter` to save the changes and quit.

## Override the API security type

The Application Registry allows you to register APIs:
- Secured with [Basic Authentication](https://tools.ietf.org/html/rfc7617)
- Secured with OAuth flow
- Secured with client certificates
- Not secured
- Protected against cross-site request forgery (CSRF) attacks

The Application Gateway calls the registered APIs accordingly, basing on the security type specified in the API registration process.

The Application Gateway overrides the registered API's security type if it gets a request which contains the **Access-Token** header. In such a case, the Application Gateway rewrites the token from the **Access-Token** header into an OAuth-compliant **Authorization** header and forwards it to the target API.

This mechanism is suited for implementations in which an external application handles user authentication.
