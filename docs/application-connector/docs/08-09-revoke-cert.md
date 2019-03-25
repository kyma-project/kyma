---
title: Revoke the client certificate
type: Tutorials
---

You can revoke a client certificate generated for your Application. Revocation prevents a certificate from being renewed. A revoked certificate, however, continues to be valid until it expires.

To revoke a client certificate, send a request to the `certificates/revocations` endpoint. Pass the certificate you want to revoke and a key that matches this certificate in the call. Run:
    
```bash
curl -X POST https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/revocations --cert {CERT_TO_REVOKE} --key {CERT_TO_REVOKE_KEY} -k 
```

If you have access to the Kyma cluster, you can also revoke a certificate by sending the SHA256 fingerprint of the certificate to the internal `certificates/revocations` endpoint.

To get SHA256 fingerprint convert certificate in the `pem` format to the `der` format:
```bash
openssl x509 -in {CLIENT_CERT_FILE}.crt -outform der -out {CLIENT_CERT_DER_FILE}.der
```
To get the SHA256 fingerprint run:
```bash
shasum -a 256 {CLIENT_CERT_DER_FORMAT}.der
```
Revoke the certificate by running:

```bash
curl -X POST http://connector-service-internal-api:8080/v1/applications/certificates/revocations -d '{hash: {CERT_TO_REVOKE_SHA256_FINGERPRINT}}'
```
