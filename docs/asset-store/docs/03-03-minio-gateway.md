---
title: Minio and Minio Gateway
type: Details
---

The whole concept of the Asset Store relies on Minio as the storage solution. It supports Kyma's manifesto and the "batteries included" rule by providing you with this on-premise solution by default.

Depending on the usage scenario, you can:
- Use Minio for local development.
- Store your assets on a production scale using Minio in a [Gateway mode](https://github.com/minio/minio/tree/master/docs/gateway).

The Asset Store ensures that both usage scenarios work for Kyma, without additional configuration of the built-in controllers.


## Development mode storage

Minio is an open-source asset storage server with Amazon S3 compatible API. You can use it to store various types of assets, such as documents, files, or images.

In the context of the Asset Store, the Asset Controller stores all assets in Minio, in a dedicated storage space.

![](./assets/minio.svg)

### Access Minio credentials

For security reasons, Minio credentials are generated during Kyma installation and stored inside the Kubernetes Secret object. To access them, run the following commands:

- Get the access key using `kubectl get secret assetstore-minio -n kyma-system -o jsonpath=“{.data.accesskey}” | base64 -D`
- Get the secret key using `kubectl get secret assetstore-minio -n kyma-system -o jsonpath=“{.data.secretkey}” | base64 -D`

You can also set Minio credentials directly using `values.yaml` files. For more details, see the official [Minio documentation](https://github.com/helm/charts/tree/master/stable/minio#configuration).


## Production storage

For the production purposes, the Asset Store uses Minio Gateway which:

- Is a multi-cloud solution that offers the flexibility to choose a given cloud provider for the specific Kyma installation, including Azure, Amazon, and Google
- Allows you to use various cloud providers that support the data replication and CDN configuration
- Is compatible with Amazon S3 APIs


![](./assets/minio-gateway.svg)

See [this tutorial](#tutorials-set-minio-to-the-google-cloud-storage-gateway-mode) to learn how to set Minio to the Google Cloud Storage Gateway mode.