---
title: Asset and Bucket CRs flow
type: Architecture
---


This diagram provides an overview of the basic Asset and Bucket CRs flow and the role of particular components in this process:

![](./assets/rafter-architecture-2.svg)

1. The user creates a bucket through a Bucket CR.
2. The Bucket Controller listens for new events and acts upon receiving the Bucket CR creation event.
3. The Bucket Controller creates the bucket in the MinIO Gateway storage.
4. The user creates an Asset CR which specifies the reference to the asset source location and the name of the bucket for storing the asset.
5. The AC listens for new events and acts upon receiving the Asset CR creation event.
6. The AC reads the CR definition, checks if the Bucket CR is available, and if its name matches the bucket name referenced in the Asset CR. It also verifies if the Bucket CR is in the `Ready` phase.
7. If the Bucket CR is available, the AC fetches the asset from the source location provided in the CR. If the asset is a ZIP or TAR file, the AC unpacks and optionally filters the asset before uploading it into the bucket.
8. Optionally, the AC validates, modifies the asset, or extracts asset's metadata if such a requirement is defined in the Asset CR. The AC communicates with the validation, mutation, and metadata services to validate, modify the asset, or extract asset's metadata according to the specification defined in the Asset CR.
9. The AC uploads the asset to MinIO Gateway, into the bucket specified in the Asset CR.
10. The AC updates the status of the Asset CR with the storage location of the file in the bucket.
