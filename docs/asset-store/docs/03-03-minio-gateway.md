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


## Production storage

For the production purposes, the Asset Store uses Minio Gateway which:

- Is a multi-cloud solution that offers the flexibility to choose a given cloud provider for the specific Kyma installation, including Azure, Amazon, and Google
- Allows you to use various cloud providers that support the data replication and CDN configuration
- Is compatible with Amazon S3 APIs


![](./assets/minio-gateway.svg)
