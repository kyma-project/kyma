---
title: Renew the client certificate
type: Tutorials
---

By default, the client certificate you generate when you connect an external solution to Kyma is valid for 92 days. Follow this tutorial to renew the client certificate.

>**NOTE:** You can only renew client certificates that are still valid. If your client certificate is expired or has been revoked, you must generate a new one.

1. To renew the client certificate, use the same certificate subject that matches the subject of your current certificate. To check the certificate subject, run:
  ```
  openssl x509 -noout -subject -in {PATH_TO_OLD_CRT}
  ```

2. Generate a new CSR using the certificate subject you got in the previous step. Run:
  ```
  openssl req -new -sha256 -out renewal.csr -key {PATH_TO_KEY} -subj "{SUBJECT}"
  ```

3. To renew a certificate, send a request to RenewCertURL obtained from management/info endpoint. Example call:

  ```
  curl -X POST https://gateway.{DOMAIN}/v1/applications/certificates/renewals -d '{"csr":"BASE64_ENCODED_CSR"}' -k --cert {PATH_TO_OLD_CRT} --key {PATH_TO_KEY}
  ```

A successful call returns a renewed client certificate:
```
{
    "crt":"BASE64_ENCODED_CRT_CHAIN",
    "clientCrt":"BASE64_ENCODED_CLIENT_CRT",
    "caCrt":"BASE64_ENCODED_CA_CRT"
}
```
