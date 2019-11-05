---
title: Revoke a client certificate
type: Tutorials
---

You can revoke a client certificate generated for your Application. Revocation prevents a certificate from being [renewed](#tutorials-renew-the-client-certificate). A revoked certificate, however, continues to be valid until it expires.

To revoke a client certificate, send a request to the `certificates/revocations` endpoint. Pass the certificate you want to revoke and a key that matches this certificate in the call. Run:
    
```bash
curl -X POST https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/revocations --cert {CERT_TO_REVOKE} --key {CERT_TO_REVOKE_KEY} -k 
```

## Revoke a certificate using the SHA256 fingerprint

If you have admin access to the Kyma cluster, you can revoke client certificates by sending the SHA256 fingerprint of a certificate to the internal `certificates/revocations` endpoint. Follow these steps: 

1. Convert the certificate from the `pem` format to the `der` format.

    ```bash
    openssl x509 -in {CLIENT_CERT_FILE}.crt -outform der -out {CLIENT_CERT_DER_FILE}.der
    ```
   
2. Get the SHA256 fingerprint of the certificate.

    ```bash
    shasum -a 256 {CLIENT_CERT_DER_FILE}.der
    ```
   
3. Revoke the certificate using the SHA256 fingerprint.

    ```bash
    curl -X POST http://connector-service-internal-api:8080/v1/applications/certificates/revocations -d '{hash: {SHA256_FINGERPRINT_OF_CERT_TO_REVOKE_}}'
    ```
