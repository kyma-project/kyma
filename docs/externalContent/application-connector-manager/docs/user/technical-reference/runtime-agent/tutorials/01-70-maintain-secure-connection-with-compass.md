# Maintain a Secure Connection with UCL

After you have established a secure connection with UCL, you can fetch the configuration details and renew the client certificate before it expires. To renew the client certificate, follow the steps in this tutorial.

## Prerequisites

- [OpenSSL toolkit](https://openssl-library.org/source/index.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards
- [UCL](https://github.com/kyma-incubator/compass) (previously called Compass)
- Registered Application
- Runtime connected to Compass
- [Established secure connection with UCL](01-60-establish-secure-connection-with-compass.md)

## Steps

1. Get the CSR information with the configuration details.

    To fetch the configuration, make a call to the Certificate-Secured Connector URL using the client certificate.
    The Certificate-Secured Connector URL is the `certificateSecuredConnectorURL` obtained when establishing a secure connection with UCL.
    Send this query with the call:

    ```graphql
    query {
        result: configuration {
            certificateSigningRequestInfo {
                subject
                keyAlgorithm
            }
            managementPlaneInfo {
                directorURL
            }
        }
    }
    ```

    A successful call returns the requested configuration details.

2. Generate a key and a Certificate Signing Request (CSR).

    Generate a CSR with this command using the certificate subject data obtained with the CSR information:

    ```bash
    export KEY_LENGTH=4096
    openssl genrsa -out ucl-app.key $KEY_LENGTH
    openssl req -new -sha256 -out ucl-app.csr -key ucl-app.key -subj "{SUBJECT}"
    ```

   > [!NOTE]
   > The key length is configurable, however, 4096 is the recommended value.

3. Sign the CSR and renew the client certificate.

    Encode the obtained CSR with base64:

    ```bash
    openssl base64 -in ucl-app.csr
    ```

    Send the following GraphQL mutation with the encoded CSR to the Certificate-Secured Connector URL:

    ```graphql
    mutation {
        result: signCertificateSigningRequest(csr: "{BASE64_ENCODED_CSR}") {
            certificateChain
            caCertificate
            clientCertificate
        }
    }
    ```

    The response contains a renewed client certificate signed by the Kyma Certificate Authority (CA), certificate chain, and the CA certificate.

4. Decode the certificate chain.

    The returned certificates and the certificate chain are base64-encoded and need to be decoded before use.
    To decode the certificate chain, run:

    ```bash
    base64 -d {CERTIFICATE_CHAIN}
    ```

> [!NOTE]
> See how to [revoke a client certificate](01-80-revoke-client-certificate.md).
