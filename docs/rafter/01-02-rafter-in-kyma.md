---
title: Rafter in Kyma
type: Overview
---

Kyma provides a Kubernetes-based solution for managing content that relies on the custom resource (CR) extensibility feature and [Rafter](/#overview-rafter) as a backend mechanism. This solution allows you to upload multiple and grouped data for a given documentation topic and store them as Asset CRs in external buckets located in [MinIO](https://min.io/) storage. All you need to do is to specify topic details, such as documentation sources, in an AssetGroup CR or a ClusterAssetGroup CR and apply it to a given Namespace or cluster. The CR supports various documentation formats, including images, Markdown documents, [AsyncAPI](https://www.asyncapi.com/), [OData](https://www.odata.org/), and [OpenAPI](https://www.openapis.org/) specification files. You can upload them as single, direct file URLs and packed assets (ZIP or TAR).

The content management solution offers these benefits:

- It provides a unified way of uploading different document types to a Kyma cluster.
- It supports baked-in documentation. Apart from the default documentation, you can add your own and group it as you like, the same way you use micro frontends to personalize views in the Console UI. For example, you can add contextual help for a given Service Broker in the Service Catalog.
