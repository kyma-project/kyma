---
title: Overview
---

The Headless CMS is a new breed of traditional Content Management Systems (CMS) that provides a way of storing and managing raw content, and exposing it through an API. It allows you to pull the content into your own application and tailor it to your needs, delivering it in any format, on any device. Contrary to the traditional CMS, such as WordPress, the Headless CMS does not provide a display layer and ready-to-use templates. Instead, it only ensures a database backend. It gives flexibility on the choice of the frontend, thus cutting the default "head" off the traditional CMS solutions.

## Headless CMS in Kyma

Kyma provides a Kubernetes-based solution that relies on the custom resource (CR) extensibility feature and the [Asset Store](/components/asset-store/#overview-overview) as a backend mechanism. The Headless CMS in Kyma allows you to upload multiple and grouped data for a given documentation topic and store them as Asset CRs in MinIO buckets. All you need to do is to specify all topic details, such as documentation sources, in a DocsTopic CR or a ClusterDocsTopic CR and apply it to a given Namespace or a cluster. The CR supports various documentation formats, including images, Markdown documents, [AsyncAPI](https://www.asyncapi.com/), [OData](https://www.odata.org/), and [OpenAPI](https://www.openapis.org/) specification files. You can upload them as single (direct file URLs) and packed assets (ZIP or TAR).

## Benefits

The Headless CMS brings a number of benefits:

- It provides a unified way of uploading different document types to a Kyma cluster.
- It fits into the Kyma modularity concept as you load onto a cluster only documentation for the installed components. This is possible as the DocsTopic CR and the code for a given component are located in the same place in the `kyma` repository.
- It supports baked-in documentation. Apart from the default documentation, you can add your own and group it as you like, the same way you use micro frontends to personalize views in the Console UI. For example, you can add contextual help for a given Service Broker in the Service Catalog.
