---
title: Connector Service
type: Architecture
---

The Connector Service generates client certificates which are used to secure the communication between Kyma and the connected external solutions.        

Generating a new client certificate is the first step in the process of configuring an Application (App). Kyma stores the root certificate and serves as the Certificate Authority when you configure a new App. When you generate a new client certificate, the Connector Service returns it along with the root certificate to allow validation.  

This diagram illustrates the client certificate generation flow in details:
![Client certificate generation operation flow](./assets/002-automatic-configuration.svg)

1. The administrator requests for a token using the CLI or the UI and receives a link with the token, which is valid for a limited period of time.
2. The administrator passes the token to the external system, which requests for information regarding the Kyma installation. In the response, it receives the following information:
    - the URL to which a third-party solution sends its Certificate Signing Request (CSR)
    - the URL of the metadata endpoint
    - information required to generate a CSR
3. The external system generates a CSR based on the information provided by Kyma and sends the CSR to the designated URL. In the response, the external system receives a signed certificate. It can use the certificate to authenticate and safely communicate with Kyma.
4. The external system calls metadata endpoint which contains the following information:
    - the URL of the Application Registry API
    - the URL of the Event Service API
    - the certificate renewal endpoint URL which can be used to generate a new certificate
    - the certificate revocation endpoint which can be used to revoke compromised certificates
    - Identity information  

>**NOTE:** Follow [this](#tutorials-get-the-client-certificate) tutorial to learn how to get a client certificate for your implementation.
