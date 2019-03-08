---
title: Revoke the client certificate
type: Tutorials
---

You can revoke a client certificate generated for your Application. Revocation prevents a certificate from being renewed. A revoked certificate, however, continues to be valid until it expires.

To revoke a client certificate, send a request to the `certificates/revocations` endpoint. Pass the certificate you want to revoke and a key that matches this certificate in the call. Run:
    
```bash
curl -X POST https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/revocations --cert {CERT_TO_REVOKE} --key {CERT_TO_REVOKE_KEY} -k 
```

If you have access to the Kyma cluster, you can also revoke a certificate by sending the SHA256 sum of the certificate to the internal `certificates/revocations` endpoint. Run:

```bash
curl -X POST http://connector-service-internal-api:8080/v1/applications/certificates/revocations -d '{hash: {CERT_TO_REVOKE_SHA256}}'
```
