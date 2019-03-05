---
title: Revoke the certificate
type: Tutorials
---

This guide will show you how to revoke certificates. Please, follow this steps:

1. Complete ``Get the client certificate`` tutorial to get infoUrl and generate client certificate.

2. Send a request to infoUrl.

    ```curl -k https://connector-service.kyma.local/v1/applications/management/info```
    
    Successfull call will return response with revokeCertUrl.
    
    ```
    {
         "clientIdentity": {
           "application": "example-application",
           "group": "example-group",
           "tenant": "example-tenant"
         },
         "urls": {
           "metadataUrl": "https://gateway.test.cluster.kyma.cx/{APP_NAME}/v1/metadata/services",
           "eventsUrl": "https://gateway.test.cluster.kyma.cx/{APP_NAME}/v1/events",
           "renewCertUrl": "https://gateway.test.cluster.kyma.cx/v1/applications/certificates/renewal",
           "revokeCertUrl": "https://gateway.test.cluster.kyma.cx/v1/applications/certificates/revocations"
         }
    }
      ```
    
3. To revoke certificate send request to revokeCertUrl.
    ```curl -X POST https://gateway.test.cluster.kyma.cx/v1/applications/certificates/revocations --cert {CERT_TO_REVOKE} --key {CERT_TO_REVOKE_KEY} -k```
    
4. To verify that certificate has been revoked, you can try to renew it by reproducing steps in tutorial ``Renew the client certificate``