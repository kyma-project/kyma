# Rafter

## Overview

Rafter is a solution for storing and managing different types of files called assets. It uses [MinIO](charts/controller-manager/charts/minio) as object storage. The whole concept of Rafter relies on Kubernetes custom resources (CRs) managed by the [Rafter Controller Manager](./charts/controller-manager). These CRs include:

- Asset CR which manages a single asset or a package of assets
- Bucket CR which manages buckets
- AssetGroup CR which manages a group of Asset CRs of a specific type to make it easier to use and extract webhook information

Rafter enables you to manage assets using supported webhooks. For example, if you use Rafter to store a file such as a specification, you can additionally define a webhook service that Rafter should call before the file is sent to storage. The webhook service can:

- validate the file
- mutate the file
- extract some of the file information and put it in the status of the custom resource

Rafter comes with the following set of services and extensions compatible with Rafter webhooks:

- [Upload Service](./charts/upload-service) (optional service)
- [Front Matter Service](./charts/front-matter-service) (extension)
- [AsyncAPI Service](./charts/asyncapi-service) (extension)

To learn more about the Rafter, go to the [Rafter repository](https://github.com/kyma-project/rafter).
