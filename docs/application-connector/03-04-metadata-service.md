---
title: Application Registry
type: Details
---

The Application Registry allows you to register the APIs and event catalogs of the services exposed by the connected external solution.         

The Application Registry stores the data of all registered services in:
- Application custom resource (CR), which stores the metadata of the service.
- ClusterAssetGroup custom resource (CR), which stores the links to API specification, event catalog, and documentation.
- Rafter Upload Service, which stores the files containing API specification, event catalog, and documentation in an Rafter bucket.
- Kubernetes Secrets, which store sensitive data, such as OAuth credentials.

## Kubernetes APIs interaction

The Application Registry interacts with Kubernetes APIs to perform these tasks:
- Modify the Application CR instance.
- Create Secrets which contain client ID and client secret used to access OAuth-secured APIs.
- Create the Access Service.
