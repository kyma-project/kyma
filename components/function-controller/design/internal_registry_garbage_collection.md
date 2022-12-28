# Internal Registry Garbage Collection


## Summary

The current implementation for the Internal Registry deployed with Kyma Functions doesn't include any garbage collection logic. In dynamic environments with several Functions with multiple versions, if the Internal Registry is used with a cluster PVC storage backend, the volume disk space can fill up very quickly, resulting in build failures for new Function versions.

In addition to the space consumed by the Function images, the Function build process pushes the build cache layers to the registry for later, faster builds.

Alternatively, if cloud storage like S3 is used as a storage backend, the old image and cache blobs will increase S3 cost overtime.

This is a proposal to implement garbage collection logic to avoid these issues.


### Goals
- Implement simple garbage collection logic to be used with the Functions Internal Registry.

### Non-goals
- Implement garbage collection for external registries.


## Proposal
The Docker registry already has a [garbage collection mechanism](https://docs.docker.com/registry/garbage-collection/). However, it's not very flexible. It's only possible to delete blobs that are not referenced by any manifests.

To utilize this, we need to implement custom logic that will periodically run and:
- Identify existing Functions running on the cluster.
- Identify current images used by those Functions' runtimes.
- List all tags for the used images.
- Identify unused tags and use the Registry API to delete them.
- List all cache layers on the registry.
- Identify unused cache layers and use the Registry API to delete corresponding tags.

One caveat with this approach is that it's possible to miss some images for Functions created and deleted between runs. It's technically possible to simply list _all_ images on the Registry and delete all images that are not currently used by a Function, but this approach has a wider blast radius. The more conservative approach is preferred.

Separately, we need to run the garbage collection tool provided by the Registry to do the actual garbage collection and remove the blobs from the file system.


## Implementation details
The garbage collection process is implemented in three phases:

### Registry API-level image garbage collection

For this phase, a simple command line tool will perform the following steps:
#### Function image garbage collection

- Identify existing Functions running on the cluster.
- Identify current images used by those Functions' runtimes.
- List all tags for the used images.
- Identify unused tags and use the Registry API to delete them.

#### Cache image garbage collection
- List all non-cache layers on the Registry. This is done after applying the Function image garbage collection to make sure only referenced images are listed.
- List all the cache layers. Each layer is mapped to its referencing tag.
- Cross check the image layers and the cached layers list. Layers referenced in both lists should be kept.
- Tags referencing the remaining cache layers are deleted using the Registry API.

This tools runs as a Kubernetes Cronjob and is deployed as part of the Internal Registry manifest.

### Unreferenced blob deletion
The simplest way to perform this is to use the `registry garbage-collection` command.

This is implemented as a simple loop in a side-car container in the registry Pod. It will run periodically and independently from the Function image garbage collection tool to remove unreferenced blobs. 

