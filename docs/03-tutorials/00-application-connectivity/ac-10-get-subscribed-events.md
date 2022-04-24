---
title: Get subscribed events
---
<!-- TODO: Adjust to the new flow -->

Eventing provides an endpoint for fetching subscribed events for the application. To fetch all of them, make a call:

> **NOTE:** `CLUSTER_DOMAIN`, `APP_NAME`, `CLIENT_CERT_FILE_NAME`, and `KEY_FILE_NAME` are the names of your cluster domain, your Application representing your external solution, and your client certificate and key generated for your Application respectively, exported as environment variables. 
>   ```bash
>   export CLUSTER_DOMAIN=local.kyma.dev
>   export APP_NAME={APP_NAME}
>   export CLIENT_CERT_FILE_NAME=generated
>   export KEY_FILE_NAME=generated
>   ```
>   <!-- TODO: Adjust to the new flow -->

> **CAUTION:** On a local Kyma deployment, skip SSL certificate verification when making a `curl` call, by adding the `-k` flag to it. Alternatively, add the Kyma certificates to your local certificate storage on your machine using the `kyma import certs` command.

```bash
curl https://gateway.$CLUSTER_DOMAIN/$APP_NAME/v1/events/subscribed --cert $CLIENT_CERT_FILE_NAME.crt --key $KEY_FILE_NAME.key
```

A successful call returns a list of all active events for the application.
