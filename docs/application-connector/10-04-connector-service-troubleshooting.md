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

This error is caused by applying wrong subject info to your Certificate Signing Request.  
To ensure it was generated properly, check the values returned by the Connector Service with the call that fetched CSR info:
```json
{
  "csrUrl":"https://connector-service.name.cluster.extend.cx.cloud.sap/v1/applications/certificates?token=5o7ucwjz9vcpFlBsHJcwnnuL-rU8af1MsfQ6OlWTgauw7aB-xtSkXUn_ts0RtMMKhvlZVPridqmAPbf2mKC8YA==",
  "api":{
    "eventsUrl":"https://gateway.name.cluster.extend.cx.cloud.sap/app/v1/events",
    "metadataUrl":"https://gateway.name.cluster.extend.cx.cloud.sap/app/v1/metadata/services",
    "infoUrl":"https://gateway.name.cluster.extend.cx.cloud.sap/v1/applications/management/info",
    "certificatesUrl":"https://connector-service.name.cluster.extend.cx.cloud.sap/v1/applications/certificates"
  },
  "certificate":{
    "subject":"O=Organization,OU=OrgUnit,L=Waldorf,ST=Waldorf,C=DE,CN=CNAME",
    "extensions":"",
    "key-algorithm":"rsa2048"
  }
}
```

Subject values present in CSR should match the subject in this response.

To check the subject of the generated CSR, run the following command:
```
openssl req -noout -subject -in {PATH_TO_CSR_FILE}
```

Any other `400` status codes may be the result of improper Base64 encoding or faulty CSR generation.


## Certificate renewal returns 403

While trying to renew a certificate you may get this error:
```json
{
  "code": 403,
  "error": "Certificate has been revoked."
}
```
It means that the certificate has been revoked. 

You cannot renew a certificate that has been revoked.  
To generate a new certificate, see [this tutorial](#tutorials-get-the-client-certificate).