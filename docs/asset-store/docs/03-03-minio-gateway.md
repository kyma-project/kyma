---
title: Minio and Minio Gateway
type: Details
---

The whole concept of the Asset Store relies on Minio as a storage solution. It supports the Kyma's manifesto and the "batteries included" rule as Kyma provides you with this on-premise solution by default.

Depending on the usage scenario, you can:
- Use Minio for local development.
- Store your assets on a production scale using Minio in a [Gateway mode](https://github.com/minio/minio/tree/master/docs/gateway).

The Asset Store ensures that both usage scenarios work for Kyma, without additional configuration of the in-build controllers.

## Local storage

Minio is an open-source asset storage server with Amazon S3 compatible API. It can be used for storing various types of assets, such as documents, images, or videos. The size of such an asset can range from a few KBs to a maximum of 5TB.

In the context of the Asset Store, the Asset Controller stores all assets in Minio, in a dedicated storage space.

![](assets/minio.svg)


## Production storage

Minio does not scale for the production use due to its limits in file sizes and HTTP requests and responses. The Gateway mode, Minio's scalable counterpart, gives you more flexibility as it allows you to use the asset storage services from such major cloud providers as Azure, Amazon, and Google. Similarly to Minio, the Gateway mode
 compatible with Amazon S3 APIs

adds Amazon S3 compatibility to the external storage.


![](assets/minio-gateway.svg)
