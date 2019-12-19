---
title: Overview
type: Overview
---

Rafter is a solution for storing and managing different types of assets, such as documents, files, images, API specifications, and client-side applications. It uses an external solution, [MinIO](https://min.io/), for storing assets. The whole concept relies on Kubernetes custom resources (CRs) managed by the Asset, Bucket, and AssetGroup controllers (and their cluster-wide counterparts) grouped under the [Rafter Controller Manager](https://github.com/kyma-project/rafter/blob/master/cmd/manager/README.md). These CRs include:

- Asset CR which manages a single asset or a package of assets
- Bucket CR which manages buckets in which these assets are stored
- AssetGroup CR which manages a group of Asset CRs of a specific type

Rafter enables you to manage assets using supported webhooks. For example, if you use Rafter to store a specification, you can additionally define a webhook service that Rafter should call before the file is sent to storage. The webhook service can:

- Validate the file
- Mutate the file
- Extract some of the file information and put it in the status of the custom resource

Rafter comes with the following set of services and extensions compatible with Rafter webhooks:

- [Upload Service](#details-upload-service) (optional service)
- [AsyncAPI Service](#details-asyncapi-service) (extension)
- [Front Matter Service](#details-front-matter-service) (extension)

## Benefits

In general, Rafter is a new breed of traditional Content Management Systems (CMS) that provides a way of storing and managing raw content and exposing it through an API. It allows you to pull the content into your own application and tailor it to your needs, delivering it in any format, on any device. Contrary to the traditional CMS, such as WordPress, Rafter does not provide a display layer and ready-to-use templates. Instead, it only ensures a database backend. It gives flexibility on the choice of the frontend thus cutting the default "head" off the traditional CMS solutions.

This solution offers a number of benefits:

- It's flexible. You can use it for storing various types of assets, such as Markdown documents, ZIP, PNG, or JS files.
- It's scalable. It allows you to store assets on a production system, using cloud provider storage services. At the same time, you can apply it to local development and use MinIO to store assets on-premise.
- It allows you to avoid vendor lock-in. When using Rafter in a production system, you can seamlessly switch between different major service providers, such as AWS S3 or Azure Blob.
- It's location-independent. It allows you to expose files directly to the Internet and replicate them to different regions. This way, you can access them easily, regardless of your location.
