# Revoke a Client Certificate (RA)

After you have established a secure connection with UCL and generated a client certificate, you may want to revoke this certificate at some point. To revoke a client certificate, follow the steps in this tutorial.

> [!NOTE]
> A revoked client certificate remains valid until it expires, but it cannot be renewed.

## Prerequisites

- [OpenSSL toolkit](https://openssl-library.org/source/index.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards
- [UCL](https://github.com/kyma-incubator/compass) (previously called Compass)
- Registered Application
- Kyma runtime connected to UCL
- [Established secure connection with UCL](01-60-establish-secure-connection-with-compass.md)

> [!NOTE]
> See how to [maintain a secure connection with UCL and renew a client certificate](01-70-maintain-secure-connection-with-compass.md).

## Revoke the Certificate

To revoke a client certificate, make a call to the Certificate-Secured Connector URL using the client certificate.
The Certificate-Secured Connector URL is the `certificateSecuredConnectorURL` obtained when establishing a secure connection with UCL.
Send this mutation with the call:

```graphql
mutation { result: revokeCertificate }
```

A successful call returns the following response:

```json
{"data":{"result":true}}
```
