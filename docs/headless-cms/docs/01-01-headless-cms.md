---
title: Overview
---

The Headless CMS is a new breed of traditional Content Management Systems (CMS) that provides a way of storing and managing raw content, and exposing it through an API. You can pull the content into your own application and tailor it to your needs, delivering it in any format, on any device. Contrary to the traditional CMS, such as WordPress, Headless CMS does not provide a display layer and ready-to-use templates. Instead, it only ensures the backend in the form of a database and gives flexibility on the choice of the frontend, thus cutting the default "head" off the traditional CMS solutions.

## Headless CMS in Kyma

Kyma takes the concept of the Headless CMS into a higher, cloud-native level. It provides a Kubernetes-based Headless CMS that relies on the custom resource (CR) extensibility feature. You add a document to the Console UI through a CR, while the backend mechanism in the form of the Asset Store handles the rest. While the Asset Store itself allows you to add separate Asset CRs and store them in respective Minio buckets, the Headless CMS ensures the upload of multiple and grouped data for a given documentation topic. All you need to do is to specify all details of the topic, including the documentation sources, in a DocsTopic CR or a ClusterDocsTopic and apply it to a given Namespace or a cluster, respectively. The CR supports various documentation formats, including images, Markdown documents, AsyncAPI, OData, and OpenAPI specification files. You can upload them both as single and packed (ZIP or TAR) assets. The labels defined in the CR specify in which view and under which group the Console UI renders the assets.

## Benefits

The Headless CMS brings a number of benefits:

- It provides a unified way of uploading different documentation types on a Kyma cluster.
- It fits the Kyma modularity concept as you load onto a cluster only documentation for the installed components. This is possible as the DocsTopic CR and code for a given component go together in the `kyma` repository.
- It supports baked-in documentation. Apart from the default documentation, you can add your own and group it as you like, the same way you personalize views in the Console UI with micro frontends. For example, you can add contextual help for a given Service Broker in the Service Catalog.
