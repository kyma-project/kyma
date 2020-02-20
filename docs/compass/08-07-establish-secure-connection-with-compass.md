---
title: Establish a secure connection with Compass
type: Tutorials
---

To establish a secure connection with Compass and generate the client certificate, follow this tutorial. 

## Prerequisites

- [OpenSSL toolkit](https://www.openssl.org/docs/man1.0.2/apps/openssl.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards
- Compass (version 1.8 or higher)
- Registered Application
- Runtime connected to Compass

## Steps

1. Get the Connector URL and the one-time token.

    To get the Connector URL and the one-time token which allow you to fetch the required configuration details, use the Compass Console.
    
    >**NOTE:** To access the Compass Console, go to the `https://compass.{CLUSTER_DOMAIN}` URL and enter your Kyma credentials. 
    
    Alternatively, make a call to the Director including the `Tenant` header with Tenant ID and `authorization` header with the Dex Bearer token. Use the following mutation: 
    
    ```graphql
    mutation { 
        result: generateOneTimeTokenForApplication(id: "{APPLICATION_ID}") { 
            token 
            connectorURL 
        }
    }
    ```
   
   > **NOTE:** The one-time token expires after 5 minutes.

2. Get the CSR information and configuration details from Kyma using the one-time token.

    To get the CSR information and configuration details, send this GraphQL query to the Connector URL.
    You must include the `connector-token` header containing the one-time token when making the call.

    ```graphql
    query {
        result: configuration {
            token {
                token
            }
            certificateSigningRequestInfo {
                subject
                keyAlgorithm
            }
            managementPlaneInfo {
                directorURL
                certificateSecuredConnectorURL
            }
        }
    }
    ``` 

    A successful call returns the data requested in the query including a new one-time token.

3. Generate a key and a Certificate Signing Request (CSR).

    Generate a CSR with the following command. `SUBJECT` is the certificate subject data returned with the CSR information as `subject`.   
    
    ```bash
    export KEY_LENGTH=4096
    openssl genrsa -out compass-app.key $KEY_LENGTH
    openssl req -new -sha256 -out compass-app.csr -key compass-app.key -subj "{SUBJECT}"
    ```
   > **NOTE:** The key length is configurable, however, 4096 is the recommended value.

4. Sign the CSR and get a client certificate. 

    Encode the obtained CSR with base64:
    ```bash
    openssl base64 -in compass-app.csr 
    ```

    To get the CSR signed, use the encoded CSR in this GraphQL mutation:
    ```graphql
    mutation {
        result: signCertificateSigningRequest(csr: "{BASE64_ENCODED_CSR}") {
            certificateChain
            caCertificate
            clientCertificate
        }
    }
    ```
   
    Send the modified GraphQL mutation to the Connector URL. You must include the `connector-token` header containing the one-time token fetched with the configuration.

    The response contains a certificate chain, a valid client certificate signed by the Kyma Certificate Authority (CA), and the CA certificate.
    
 5. Decode the certificate chain.
 
    After you receive the certificates, decode the certificate chain with the base64 method and use it in your application: 
    ```bash
    base64 -d {CERTIFICATE_CHAIN}
    ```
    
 >**NOTE:** To learn how to renew a client certificate, read [this](#tutorials-maintain-a-secure-connection-with-compass) tutorial.