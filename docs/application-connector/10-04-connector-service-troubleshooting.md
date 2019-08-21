---
title: Errors when generating or renewing a certificate
type: Troubleshooting
---

## Invalid Certificate Signing Request (CSR)

If you try to generate a client certificate, you may get this error:
```json
{
  "code":403,
  "error":"CSR: Invalid common name provided."
}
```

This error is caused by applying wrong subject information to your Certificate Signing Request.  
To ensure CSR was generated properly, check the values returned by the Connector Service with the call that fetched CSR information:
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

Subject values present in CSR should match the subject in the response that you got.

To check the subject of the generated CSR, run the following command:
```
openssl req -noout -subject -in {PATH_TO_CSR_FILE}
```
