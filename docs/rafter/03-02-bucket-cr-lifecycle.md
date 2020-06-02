---
title: Bucket custom resource lifecycle
type: Details
---

Learn about the lifecycle of the Bucket custom resource (CR) and how its creation and removal affect other Rafter components.

>**NOTE:** This lifecycle also applies to the ClusterBucket CR.

## Create a Bucket CR

When you create a Bucket CR, the Bucket Controller (BC) receives a CR creation Event and creates a bucket with the name specified in the CR. It is created in the MinIO Gateway storage under the `{CR_name}-{ID}` location, such as `test-bucket-1b19rnbuc6ir8`, where `{CR_name}` is the **name** field from the Bucket CR and `{ID}` is a randomly generated string. The status of the CR contains a reference URL to the created bucket.

![Create a bucket](./assets/create-bucket.svg)

## Remove a Bucket CR

When you remove the Bucket CR, the BC receives a CR deletion Event and removes the bucket with the whole content from MinIO Gateway.

The Asset Controller (AC) also monitors the status of the referenced bucket. The AC checks the Bucket CR status to make sure the bucket exists. If you delete the bucket, the AC receives information that the files are no longer accessible and the bucket was removed. The AC updates the status of the Asset CR to `ready: False` and removes the asset storage reference. The Asset CR is still available and you can use it later for a new bucket.

![Delete a bucket](./assets/delete-bucket.svg)
