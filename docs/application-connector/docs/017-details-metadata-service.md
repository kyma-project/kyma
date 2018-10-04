---
title: Metadata Service
type: Details
---

The Metadata Service allows you to register the APIs and Event catalogs of the services exposed by the connected external solution.         

## Implementation Details

### Data storage

Service's data is stored in:
- [Remote Environment CRD](https://github.com/kyma-project/kyma/blob/master/docs/application-connector/docs/040-cr-remote-evironment.md)
- [Minio](https://minio.io/)
- Kubernetes secrets

Remote Environment CRD contains registered APIs and OAuth server urls.

Minio storage contains:
- API specification
- Events catalog
- Documentation

### Kubernetes APIs usage

Metadata Service interacts with Kubernetes APIs to perform the following tasks:
- Modifying Remote Environment CRD instance
- Creating secrets containing client id and client secret used to access OAuth secured APIs
- Creating service used to access Proxy Service from lambda or service deployed on Kyma  
