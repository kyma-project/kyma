---
title: Revoke a client certificate (AC)
---

You can revoke a client certificate generated for your Application. Revocation prevents a certificate from being [renewed](ac-06-renew-client-cert.md). A revoked certificate, however, continues to be valid until it expires. 

## Revoke the certificate

1. Export the names of the generated client certificate and key, and your [cluster domain](../../02-get-started/01-quick-install.md#export-your-cluster-domain) as environment variables:

   ```bash
   export CLIENT_CERT_FILE_NAME=generated
   export KEY_FILE_NAME=generated
   export CLUSTER_DOMAIN=local.kyma.dev
   ```

2. To revoke a client certificate, send a request to the `certificates/revocations` endpoint. Pass the certificate you want to revoke and the key that matches this certificate in the call. Run:
    
   > **CAUTION:** On a local Kyma deployment, skip SSL certificate verification when making a `curl` call, by adding the `-k` flag to it. Alternatively, add the Kyma certificates to your local certificate storage on your machine using the `kyma import certs` command.

   ```bash
   curl -X POST https://gateway.$CLUSTER_DOMAIN/v1/applications/certificates/revocations --cert $CLIENT_CERT_FILE_NAME.crt --key $KEY_FILE_NAME.key 
   ```

## Revoke a certificate using the SHA256 fingerprint

If you have admin access to the Kyma cluster, you can revoke client certificates by sending the SHA256 fingerprint of a certificate to the internal `certificates/revocations` endpoint. Follow these steps: 

1. Convert the certificate from the `pem` format to the `der` format.

    ```bash
    openssl x509 -in {CLIENT_CERT_FILE_NAME}.crt -outform der -out {CLIENT_CERT_DER_FILE_NAME}.der
    ```
   
2. Get the SHA256 fingerprint of the certificate.

    ```bash
    shasum -a 256 {CLIENT_CERT_DER_FILE_NAME}.der
    ```
   
3. Revoke the certificate using the SHA256 fingerprint.

    ```bash
    curl -X POST http://connector-service-internal-api:8080/v1/applications/certificates/revocations -d '{hash: {SHA256_FINGERPRINT_OF_CERT_TO_REVOKE_}}'
    ```
