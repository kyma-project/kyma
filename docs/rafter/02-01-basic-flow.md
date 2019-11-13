---
title: Basic content flow
type: Architecture
---

## Resources

The whole concept of Rafter relies on the following components:

- **Asset custom resource** (CR) is an obligatory [CR](#custom-resource-asset) in which you define the asset you want to store in a given storage bucket. Its definition requires the asset name and mode, the name of the Namespace in which it is available, the address of its web location, and the name of the bucket in which you want to store it. Optionally, you can specify the validation and mutation requirements that the asset must meet before it is stored.

- **Asset Controller** (AC) manages the [Asset CR lifecycle](#details-asset-custom-resource-lifecycle).

- **AssetGroup custom resource** (CR) orchestrates the creation of multiple Asset CRs for a specific documentation topic in a given Namespace.

- **AssetGroup Controller** creates Asset custom resources (CRs) based on an AssetGroup CR definition. If the AssetGroup CR defines two sources of documentation topics, such as `asyncapi` and `markdown`, the AssetGroup Controller creates two Asset CRs. The AssetGroup Controller also monitors the status of the Asset CR and updates the status of the AssetGroup CR accordingly.

- **Bucket CR** is an obligatory [CR](#custom-resource-bucket) in which you define the name of the bucket for storing assets.

- **Bucket Controller** manages the [Bucket CR lifecycle](#details-bucket-custom-resource-lifecycle).

- **Validation Service** is an optional service which ensures that the asset meets the validation requirements specified in the Asset CR before uploading it to the bucket. The service returns the validation status to the AC.

- **Mutation Service** is an optional service which ensures that the asset is modified according to the mutation specification defined in the Asset CR before it is uploaded to the bucket. The service returns the modified asset to the AC.

- [**Front Matter Service**](#details-front-matter-service) is an optional service which extracts metadata from assets. The metadata information is stored in the CR status. The service returns the asset metadata to the AC.

- **MinIO Gateway** is a MinIO cluster mode which is a production-scalable storage solution. It ensures flexibility of using asset storage services from major cloud providers, including Azure Blob Storage, Amazon S3, and Google Cloud Storage.

>**NOTE:** All CRs and controllers have their cluster-wide counterparts, names of which start with the **Cluster** prefix, such as ClusterAssetGroup CR.

## Basic content flow

This diagram shows a high-level overview of how Rafter works:

>**NOTE:** This flow also applies to the cluster-wide counterparts of all CRs.

![](./assets/basic-architecture.svg)

1. Create an AssetGroup CR that manages a single asset or a package of assets you want to upload within a specified category.
2. Rafter creates Asset CRs in the number specified in the AssetGroup CR.
3. Services implemented for Rafter webhooks, can optionally validate, mutate, or extract data from assets before uploading them into buckets.
4. Rafter creates Bucket CRs that define buckets in which assets must be stored.
5. Rafter creates buckets in Minio and moves assets into them.
