---
title: Maintain a secure connection with Compass
type: Tutorials
---

After you have established a secure connection with Compass, you can fetch the configuration details and renew the client certificate before it expires. To renew the client certificate, follow the steps in this tutorial.

## Prerequisites

- [OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards
- Compass (version 1.8 or higher)
- Registered Application
- Runtime connected to Compass
- [Established secure connection with Compass](#tutorials-establish-a-secure-connection-with-compass)

## Steps

1. Get the CSR information with the configuration details.

    To fetch the configuration, make a call to the Certificate-Secured Connector URL using the client certificate. 
    The Certificate-Secured Connector URL is the `certificateSecuredConnectorURL` obtained when establishing a secure connection with Compass. 
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
    ```
    export KEY_LENGTH=4096
    openssl genrsa -out compass-app.key $KEY_LENGTH
    openssl req -new -sha256 -out compass-app.csr -key compass-app.key -subj "{SUBJECT}"
    ```
   > **NOTE:** The key length is configurable, however, 4096 is the recommended value.

3. Sign the CSR and renew the client certificate. 

    Encode the obtained CSR with base64:
    ```bash
    openssl base64 -in compass-app.csr 
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
    
>**NOTE:** To learn how to revoke a client certificate, read [this](#tutorials-revoke-a-client-certificate) tutorial.
