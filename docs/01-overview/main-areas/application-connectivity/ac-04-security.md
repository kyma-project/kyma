---
title: Security
---

## Client certificates

To provide maximum security, Application Connector uses the TLS protocol with Client Authentication enabled. As a result, whoever wants to connect to Application Connector must present a valid client certificate, which is dedicated to a specific Application. In this way, the traffic is fully encrypted and the client has a valid identity.

## TLS certificate verification

By default, the TLS certificate verification is enabled when sending data and requests to every application.
You can [disable the TLS certificate verification] (../../../03-tutorials/00-application-connectivity/ac-11-disable-tls-certificate-verification.md) in the communication between Kyma and an application to allow Kyma to send requests and data to an unsecured application. Disabling the certificate verification can be useful in certain testing scenarios.

## API security type

Application Registry allows you to register APIs:
- Secured with [Basic Authentication](https://tools.ietf.org/html/rfc7617)
- Secured with OAuth flow
- Secured with client certificates
- Not secured
- Protected against cross-site request forgery (CSRF) attacks

Application Gateway calls the registered APIs accordingly, basing on the security type specified in the API registration process.

Application Gateway overrides the registered API's security type if it gets a request which contains the **Access-Token** header. In such a case, Application Gateway rewrites the token from the **Access-Token** header into an OAuth-compliant **Authorization** header and forwards it to the target API.

This mechanism is suited for implementations in which an external application handles user authentication.
