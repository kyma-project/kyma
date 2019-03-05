---
title: Revoke the certificate
type: Tutorials
---

This guide will show you how to revoke certificates. Please, follow this steps:

1. Complete ``Get the client certificate`` tutorial to get infoUrl and generate client certificate.

2. Send a request to infoUrl.

    ```curl -k https://gateway.{CLUSTER_DOMAIN}/v1/applications/management/info```
    
    Successful call will return response with revokeCertUrl.
    
    ```json
    {
         "clientIdentity": {
           "application": "example-application",
           "group": "example-group",
           "tenant": "example-tenant"
         },
         "urls": {
           "metadataUrl": "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/metadata/services",
           "eventsUrl": "https://gateway.{CLUSTER_DOMAIN}/{APP_NAME}/v1/events",
           "renewCertUrl": "https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/renewal",
           "revokeCertUrl": "https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/revocations"
         }
    }
      ```
    
3. To revoke certificate send request to revokeCertUrl.
    ```curl -X POST https://gateway.{CLUSTER_DOMAIN}/v1/applications/certificates/revocations --cert {CERT_TO_REVOKE} --key {CERT_TO_REVOKE_KEY} -k```
    
    After successful call, the certificate is revoked and cannot be renewed.