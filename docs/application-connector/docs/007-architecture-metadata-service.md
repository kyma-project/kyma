---
title: Metadata Service
type: Details
---

## Overview

Metadata Service is responsible for registering APIs and events exposed by an external system.         

# API

Metadata Service exposes an API built around the concept of service describing external system capabilities.    
Service contains the following:
- API definition comprised of urls and specification following [OpenAPI 2.0 standard](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md).
- Events catalog following [Asynchronous API standard](https://github.com/asyncapi/asyncapi/blob/develop/schema/asyncapi.json).
- Documentation

Service definition may contain both API and Events Catalog definition or only one of those. Providing documentation is optional.  

Metadata Service supports registering APIs secured with OAuth - the user can specify OAuth server url along with client id and client secret.

It is possible to register many services for particular Remote Environment. The Metadata Service API offers a great deal of flexibility as the user may decide how they want to expose their APIs.      

Metadata Service API is described [here](https://github.com/kyma-project/kyma/blob/master/docs/application-connector/docs/assets/metadataapi.yaml).
For a complete information on registering services, please see Managing Registered Services with Metadata API Guide.


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

