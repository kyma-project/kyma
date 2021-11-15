---
title: Application Connectivity - useful links
---

If you're interested in learning more about the Application Connectivity area, follow these links to:

- Perform some simple and more advanced tasks:

  - [Pass the access token in the request header](../../../04-operation-guides/operations/ac-01-pass-access-token-in-request-header.md)
  - [Register an API in Application Registry](../../../04-operation-guides/operations/ac-02-api-registration.md)
  - [Provide a custom AC certificate and key](../../../04-operation-guides/operations/ac-03-application-connector-certificates.md)
  - [Create a new Application](../../../03-tutorials/00-application-connectivity/ac-01-create-application.md)
  - [Get the client certificate](../../../03-tutorials/00-application-connectivity/ac-02-get-client-certificate.md)
  - [Register a service](../../../03-tutorials/00-application-connectivity/ac-03-register-manage-services.md)
  - [Register a secured API](../../../03-tutorials/00-application-connectivity/ac-04-register-secured-api.md)
  - [Call a registered external service from Kyma](../../../03-tutorials/00-application-connectivity/ac-05-call-registered-service-from-kyma.md)
  - [Renew a client certificate](../../../03-tutorials/00-application-connectivity/ac-06-renew-client-cert.md)
  - [Revoke a client certificate](../../../03-tutorials/00-application-connectivity/ac-07-revoke-client-cert.md)
  - [Rotate the Root certificate and the key issued by the Certificate Authority](../../../03-tutorials/00-application-connectivity/ac-08-rotate-root-ca.md)
  - [Get the API specification for AC components](../../../03-tutorials/00-application-connectivity/ac-09-get-api-specification.md)
  - [Get subscribed events](../../../03-tutorials/00-application-connectivity/ac-10-get-subscribed-events.md)
  - [Disable TLS certificate verification](../../../03-tutorials/00-application-connectivity/ac-11-disable-tls-certificate-verification.md)

- Check Application Connectivity troubleshooting guides for:

  - [Application Gateway](../../../04-operation-guides/troubleshooting/ac-01-application-gateway-troubleshooting.md), for errors when calling a registered service
  - [Application Registry](../../../04-operation-guides/troubleshooting/ac-02-application-registry-troubleshooting.md), for certificate-related errors when trying to access the component
  - [Connector Service](../../../04-operation-guides/troubleshooting/ac-03-connector-service-troubleshooting.md), for errors when trying to renew or rotate a certificate

- Analyze Application Connectivity specification and configuration files:

  - [Application](../../../05-technical-reference/00-custom-resources/ac-01-application.md) custom resource (CR)
  - [TokenRequest](../../../05-technical-reference/00-custom-resources/ac-02-tokenrequest.md) CR
  - [Application Connector chart](../../../05-technical-reference/00-configuration-parameters/ac-01-application-connector-chart.md)
  - [Application Registry sub-chart](../../../05-technical-reference/00-configuration-parameters/ac-02-application-registry-sub-chart.md)
  - [Connector Service sub-chart](../../../05-technical-reference/00-configuration-parameters/ac-03-connector-service-sub-chart.md)
  - [Application Connectivity Certs Setup Job](../../../05-technical-reference/00-configuration-parameters/ac-04-application-connectivity-certs-setup-job.md)

- Understand technicalities behind the Application Connectivity implementation:

  - [Application Connector components](../../../05-technical-reference/00-architecture/ac-01-application-connector-components.md)
  - [Connector Service workflow](../../../05-technical-reference/00-architecture/ac-02-connector-service.md)
  - [Application Gateway workflow](../../../05-technical-reference/00-architecture/ac-03-application-gateway.md)
  - [Application Gateway details](../../../05-technical-reference/ac-01-application-gateway-details.md)
