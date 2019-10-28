---
title: Architecture
---

## Resources

The whole concept of the Asset Store relies on the following components:

- **Asset custom resource** (CR) is an obligatory [CR](#custom-resource-asset) in which you define the asset you want to store in a given storage bucket. Its definition requires the asset name and mode, the name of the Namespace in which it is available, the address of its web location, and the name of the bucket in which you want to store it. Optionally, you can specify the validation and mutation requirements that the asset must meet before it is stored.

- **Asset Controller** (AC) manages the [Asset CR lifecycle](#details-asset-custom-resource-lifecycle).

- **Bucket CR** is an obligatory [CR](#custom-resource-bucket) in which you define the name of the bucket for storing assets.

- **Bucket Controller** manages the [Bucket CR lifecycle](#details-bucket-custom-resource-lifecycle).

- **Validation Service** is an optional service which ensures that the asset meets the validation requirements specified in the Asset CR before uploading it to the bucket. The service returns the validation status to the AC.

- **Mutation Service** is an optional service which ensures that the asset is modified according to the mutation specification defined in the Asset CR before it is uploaded to the bucket. The service returns the modified asset to the AC.

- [**Metadata Service**](#details-asset-metadata-service) is an optional service which extracts metadata from assets. The metadata information is stored in the CR status. The service returns the asset metadata to the AC.

- **MinIO Gateway** is a MinIO cluster mode which is a production-scalable storage solution. It ensures flexibility of using asset storage services from major cloud providers, including Azure Blob Storage, Amazon S3, and Google Cloud Storage.

## Asset flow

This diagram provides an overview of the basic Asset Store workflow and the role of particular components in this process:

![](./assets/asset-store-architecture.svg)

1. The Kyma user creates a bucket through a Bucket CR.
2. The Bucket Controller listens for new Events and acts upon receiving the Bucket CR creation Event.
3. The Bucket Controller creates the bucket in the MinIO Gateway storage.
4. The Kyma user creates an Asset CR which specifies the reference to the asset source location and the name of the bucket for storing the asset.
5. The AC listens for new Events and acts upon receiving the Asset CR creation Event.
6. The AC reads the CR definition, checks if the Bucket CR is available, and if its name matches the bucket name referenced in the Asset CR. It also verifies if the Bucket CR is in the `Ready` phase.
7. If the Bucket CR is available, the AC fetches the asset from the source location provided in the CR. If the asset is a ZIP or TAR file, the AC unpacks and optionally filters the asset before uploading it into the bucket.
8. Optionally, the AC validates, modifies the asset, or extracts asset's metadata if such a requirement is defined in the Asset CR. The AC communicates with the validation, mutation, and metadata services to validate, modify the asset, or extract asset's metadata according to the specification defined in the Asset CR.
9. The AC uploads the asset to MinIO Gateway, into the bucket specified in the Asset CR.
10. The AC updates the status of the Asset CR with the storage location of the file in the bucket.
