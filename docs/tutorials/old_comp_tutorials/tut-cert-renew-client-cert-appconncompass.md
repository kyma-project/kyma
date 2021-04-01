---
title: Renew a client certificate
type: Tutorials
---

By default, a client certificate you generate when you connect an external solution to Kyma is valid for 92 days. Follow this tutorial to renew a client certificate.

>**NOTE:** You can only renew client certificates that are still valid. If your client certificate is expired or revoked, you must generate a new one.

1. To renew a client certificate, use the certificate subject that matches the subject of your current certificate. To check the certificate subject, run:

   ```bash
   openssl x509 -noout -subject -in {PATH_TO_OLD_CRT}
   ```

2. Generate a new Certificate Signing Request (CSR) using the certificate subject you got in the previous step.

   ```bash
   openssl req -new -sha256 -out renewal.csr -key {PATH_TO_KEY} -subj "{SUBJECT}"
   ```

3. Send a request to the Connector Service to renew the certificate.

   ```bash
   curl -X POST https://gateway.{DOMAIN}/v1/applications/certificates/renewals -d '{"csr":"BASE64_ENCODED_CSR"}' -k --cert {PATH_TO_OLD_CRT} --key {PATH_TO_KEY}
   ```

   A successful call returns a renewed client certificate:

   ```json
   {
       "crt":"BASE64_ENCODED_CRT_CHAIN",
       "clientCrt":"BASE64_ENCODED_CLIENT_CRT",
       "caCrt":"BASE64_ENCODED_CA_CRT"
   }
   ```
