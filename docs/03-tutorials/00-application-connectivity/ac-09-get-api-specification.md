---
title: Get the API specification for AC components
---

To view or download the API specification directly from this website, see the **API Consoles** section. Alternatively, get it by calling the API directly. To do so, follow the instructions in this tutorial. 

## Connector Service API

To get the API specification for the Connector Service on a local Kyma deployment, run this command:

```bash
curl https://connector-service.local.kyma.dev/v1/api.yaml
```

Alternatively, get the API specification directly from the Connector Service: 

```bash
https://connector-service.{CLUSTER_DOMAIN}/v1/api.yaml
```

## Application Registry API

To get the API specification for Application Registry for a given version of the service, run this command:

> **NOTE:** `CLUSTER_DOMAIN`, `APP_NAME`, `CLIENT_CERT_FILE_NAME`, and `KEY_FILE_NAME` are the names of your cluster domain, your Application representing your external solution, and your client certificate and key generated for your Application respectively, [exported as environment variables](ac-02-get-client-certificate.md#generate-a-csr-and-send-it-to-kyma).

```bash
curl https://gateway.$CLUSTER_DOMAIN/$APP_NAME/v1/metadata/api.yaml -k --cert $CLIENT_CERT_FILE_NAME.crt --key KEY_FILE_NAME.crt
```

