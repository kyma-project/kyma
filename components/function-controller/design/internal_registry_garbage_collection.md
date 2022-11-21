# Internal Registry Garbage Collection


## Summary

The current implementation for the Internal Registry deployed with Kyma Functions doesn't include any garbage collection logic. In dynamic environments with several Functions with multiple versions,  if the Internal Registry is used with a cluster PVC storage backend, the volume disk space can fill up very quickly resulting in build failures for new Function versions.

Alternatively, if cloud storage like is S3 is used as a storage backend, the old image blobs will will increase S3 cost overtime.

This is a proposal to implement garbage collection logic to avoid these issues.


### Goals
- Implement simple garbage collection logic to be used with the Functions Internal Registry.

### Non-goals
- Implement garbage collection for external registries.


## Proposal
The Docker registry already has a garbage collection mechanism https://docs.docker.com/registry/garbage-collection/ . However, it's not very flexible. It's only possible to delete blobs that are not referenced by any manifests.

To utilize this, we need to implement custom logic that will periodically run and:
- Identify existing functions running on the cluster.
- Identify current images used by those functions' runtimes.
- List all tags for the used images.
- Identify unused tags and use the Registry API to delete them.

One caveat with this approach is that it's possible to miss some images for functions created and deleted between runs. It's technically possible to simply list _all_ images on the Registry and delete all images that are not currently used by a function, but this approach has a wider blast radius. It is preferred the more conservative approach.

Separately, we need to run the garbage collection tool provided by the Registry to do the actual garbage collection and remove the blobs from the file system.


## Implementation details
The garbage collection process is implemented in two phases:

### Function image garbage collection
For this phase, a simple command line tool will perform the following steps:

- Identify existing functions running on the cluster.
- Identify current images used by those functions' runtimes.
- List all tags for the used images.
- Identify unused tags and use the Registry API to delete them

This tools runs as a Kubernetes Cronjob and is deployed as part of the Internal Registry manifest.

### Unreferenced blob deletion
The simplest way to perform this is to use the `registry garbage-collection` command.

This is implemented as simple loop in a side-car container in the registry pod. It will run periodically and independently from the Function image garbage collection tool to remove unreferenced blobs. 

