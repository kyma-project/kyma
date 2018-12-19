---
title: Application Registry
type: Details
---

The Application Registry allows you to register the APIs and Event catalogs of the services exposed by the connected external solution.         

The Application Registry stores the data of all registered services in:
- Application custom resource (CR), which stores the metadata of the service.
- Minio bucket, which stores the API specification, Event catalog and documentation.
- Kubernetes secrets, which stores sensitive data, such as OAuth credentials.

## Kubernetes APIs interaction

The Application Registry interacts with Kubernetes APIs to perform these tasks:
- Modify the Application CR instance.
- Create Secrets which contain client ID and client secret used to access OAuth-secured APIs.
- Create the Access Service.