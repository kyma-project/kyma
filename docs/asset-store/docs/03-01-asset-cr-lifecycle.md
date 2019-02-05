---
title: Asset custom resource lifecycle
type: Details
---

Read about the lifecycle of the Asset custom resource (CR) and how other resources react to its creation, removal, or the change in the bucket reference.

## Create an Asset CR

When you create an Asset CR, the Asset Controller receives an Event on its creation, reads its definition, verifies if the bucket exists, downloads the asset, unpacks it, and stores it in Minio Gateway.

![](assets/create-asset.svg)

## Remove an Asset CR

When you remove the Asset CR, the Asset Controller receives an Event on its deletion and deletes it from Minio Gateway.

![](assets/delete-asset.svg)

## Change the bucket reference

When you modify an Asset CR by updating the bucket reference in the Asset CR to a new one, and the previous bucket still exists, the lifecycle starts again. The asset is created in a new storage location and its new location is updated in the Asset CR.

Unfortunately, this causes duplication of data as the assets from the previous bucket storage are not cleaned up by default. Thus, to avoid multiplication of assets, first remove the old Bucket CR and then modify the existing Asset CR with the new bucket reference.

![](assets/modify-asset.svg)

## Other changes

The Asset Controller does not act upon changes in the Asset CR other than the change of the bucket reference. Thus, if the location of the source files or the **mode** of the files changes, you need to remove the old Asset CD and create a new one with the correct details.
