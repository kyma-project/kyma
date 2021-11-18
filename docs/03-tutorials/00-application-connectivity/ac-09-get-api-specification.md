---
title: Get the API specification for AC components
---

Get the API specification by calling the API directly. To do so, follow the instructions in this tutorial. 

> **NOTE:** `CLUSTER_DOMAIN`, `APP_NAME`, `CLIENT_CERT_FILE_NAME`, and `KEY_FILE_NAME` are the names of your cluster domain, your Application representing your external solution, and your client certificate and key generated for your Application respectively, [exported as environment variables](ac-02-get-client-certificate.md#generate-a-csr-and-send-it-to-kyma).

> **CAUTION:** On a local Kyma deployment, skip SSL certificate verification when making a `curl` call, by adding the `-k` flag to it. Alternatively, add the Kyma certificates to your local certificate storage on your machine using the `kyma import certs` command.

## Connector Service API

To get the API specification for Connector Service, run this command:

```bash
curl https://connector-service.$CLUSTER_DOMAIN/v1/api.yaml
```

## Application Registry API

To get the API specification for Application Registry for a given version of the service, run this command:

```bash
curl https://gateway.$CLUSTER_DOMAIN/$APP_NAME/v1/metadata/api.yaml --cert $CLIENT_CERT_FILE_NAME.crt --key $KEY_FILE_NAME.key
```

