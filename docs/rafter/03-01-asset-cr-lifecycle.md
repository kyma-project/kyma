---
title: Asset custom resource lifecycle
type: Details
---

Learn about the lifecycle of the Asset custom resource (CR) and how its creation, removal, or a change in the bucket reference affects other Rafter components.

>**NOTE:** This lifecycle also applies to the ClusterAsset CR.

## Create an Asset CR

When you create an Asset CR, the Asset Controller (AC) receives a CR creation Event, reads the CR definition, verifies if the bucket exists, downloads the asset, unpacks it, and stores it in MinIO Gateway.

![](./assets/create-asset.svg)

## Remove an Asset CR

When you remove the Asset CR, the AC receives a CR deletion Event and deletes the CR from MinIO Gateway.

![](./assets/delete-asset.svg)

## Change the bucket reference

When you modify an Asset CR by updating the bucket reference in the Asset CR to a new one while the previous bucket still exists, the lifecycle starts again. The asset is created in a new storage location and this location is updated in the Asset CR.

Unfortunately, this causes duplication of data as the assets from the previous bucket storage are not cleaned up by default. Thus, to avoid multiplication of assets, first remove one Bucket CR and then modify the existing Asset CR with a new bucket reference.

![](./assets/modify-bucket-ref-asset.svg)

## Change the Asset CR specification

When you modify the Asset CR specification, the lifecycle starts again. The previous asset content is removed and no longer available.

![](./assets/modify-asset.svg)
