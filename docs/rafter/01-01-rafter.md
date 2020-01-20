---
title: Overview
type: Overview
---

Rafter is a solution for storing and managing different types of public assets, such as documents, files, images, API specifications, and client-side applications. It uses an external solution, [MinIO](https://min.io/), for storing assets. The whole concept relies on [Kubernetes custom resources (CRs)](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) managed by the Asset, Bucket, and AssetGroup controllers (and their cluster-wide counterparts) grouped under the [Rafter Controller Manager](https://github.com/kyma-project/rafter/blob/master/cmd/manager/README.md). These CRs include:

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

>**CAUTION:** Rafter does not enforce any access control. To protect the confidentiality of your information, use Rafter only to store public data. Do not use it to process and store any kind of confidential information, including personal data.

## What Rafter is not

* Rafter is not a Wordpress-like [Content Management System](https://en.wikipedia.org/wiki/Content_management_system).
* Rafter is not a solution for [Enterprise Content Management](https://en.wikipedia.org/wiki/Enterprise_content_management).
* Rafter doesn't come with any out-of-the-box UI that allows you to modify or consume files managed by Rafter.

## What Rafter can be used for

* Rafter is based on [CRs](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/). Therefore, it is an extension of Kubernetes API and should be used mainly by developers building their solutions on top of Kubernetes,
* Rafter is a file store that allows you to programmatically modify, validate the files and/or extract their metadata before they go to storage. Content of those files can be fetched using an API. This is a basic functionality of the [headless CMS](https://en.wikipedia.org/wiki/Headless_content_management_system) concept. If you want to deploy an application to Kubernetes and enrich it with additional documentation or specifications, you can do it using Rafter,
* Rafter is an S3-like file store also for files written in HTML, CSS, and JS. It means that Rafter can be used as a hosting solution for client-side applications.

## Benefits

This solution offers a number of benefits:

- It's flexible. You can use it for storing various types of assets, such as Markdown documents, ZIP, PNG, or JS files.
- It's scalable. It allows you to store assets on a production system, using cloud provider storage services. At the same time, you can apply it to local development and use MinIO to store assets on-premise.
- It allows you to avoid vendor lock-in. When using Rafter in a production system, you can seamlessly switch between different major service providers, such as AWS S3 or Azure Blob.
- It's location-independent. It allows you to expose files directly to the Internet and replicate them to different regions. This way, you can access them easily, regardless of your location.
