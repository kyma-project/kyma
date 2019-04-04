---
title: Application Registry
type: Details
---

The Application Registry allows you to register the APIs and Event catalogs of the services exposed by the connected external solution.         

The Application Registry stores the data of all registered services in:
- Application custom resource (CR), which stores the metadata of the service.
- Docs Topic Custom Resource (CR), which stores the links to API specification, Event catalog, and documentation.
- Upload Service, which stores the files containing API specification, Event catalog, and documentation in an Asset Store bucket.
- Kubernetes secrets, which stores sensitive data, such as OAuth credentials.

## Kubernetes APIs interaction

The Application Registry interacts with Kubernetes APIs to perform these tasks:
- Modify the Application CR instance.
- Create Secrets which contain client ID and client secret used to access OAuth-secured APIs.
- Create the Access Service.