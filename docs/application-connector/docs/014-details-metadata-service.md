---
title: Metadata Service
type: Details
---

The Metadata Service allows you to register the APIs and Event catalogs of the services exposed by the connected external solution.         

The Metadata Service stores the data of all registered services in:
- Remote Environment custom resource (CR), which stores the metadata of the service.
- Minio bucket, which stores the API specification, Event catalog and documentation.
- Kubernetes secrets, which stores sensitive data, such as OAuth credentials.

## Kubernetes APIs interaction

The Metadata Service interacts with Kubernetes APIs to perform these tasks:
- Modify the Remote Environment CR instance.
- Create Secrets which contain client ID and client secret used to access OAuth-secured APIs.
- Create the Access Service.
