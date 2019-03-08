---
title: Revoke the certificate
type: Tutorials
---

This tutorial will show you how to revoke certificate, making it unable to be renewed.

To revoke certificate make a call to ``certificates/revocations`` endpoint:
    
```bash
curl -X POST https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/revocations --cert {CERT_TO_REVOKE} --key {CERT_TO_REVOKE_KEY} -k 
```
    
After successful call, the certificate is revoked and cannot be renewed.

If you have access to Kyma cluster, you can also revoke a certificate by making a call with certificates sha256 sum to internal ``certificates/revocations`` endpoint:

```bash
curl -X POST http://connector-service-internal-api:8080/v1/applications/certificates/revocations -d '{hash: {CERT_TO_REVOKE_SHA256}}'
```