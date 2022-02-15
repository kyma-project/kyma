---
title: Errors when generating or renewing a certificate
---

## Invalid Certificate Signing Request (CSR)

### Symptom

You try to generate or renew a client certificate and get this error:

```json
{
  "code":403,
  "error":"CSR: Invalid common name provided."
}
```

### Cause

This error is caused by applying wrong subject information to your Certificate Signing Request.

### Remedy

To ensure CSR was generated properly, check the values returned by Connector Service with the call that fetched the CSR information:

```json
{
  ...
  "certificate":{
    "subject":"O=Organization,OU=OrgUnit,L=Waldorf,ST=Waldorf,C=DE,CN=CNAME",
    "extensions":"",
    "key-algorithm":"rsa2048"
  }
}
```

Subject values present in the CSR should match the subject in the response that you got.

To check the subject of the generated CSR, run this command:

```bash
openssl req -noout -subject -in {PATH_TO_CSR_FILE}
```