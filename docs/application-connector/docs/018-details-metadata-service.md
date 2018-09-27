---
title: Connector Service
type: Details
---

## Overview
Metadata Service is responsible for registering APIs and events exposed by an external system.         
    

## Basic concepts 

APIs or events exposed by the external system need to be registered in Metadata Service so that they can be used in lambdas and services deployed on Kyma. 
Metadata Service exposes an API built around the concept of service describing external system capabilities.    
Service contains the following:
- API definition in a form of [OpenAPI specification](TODO)
- Events catalog in a form of [Asynchronous API specification](TODO)
- Documentation 

Service may specify:
- API
- Events Catalog
- Both API and Events catalog


Metadata Service supports registering APIs secured with OAuth ; the user can specify OAuth server url along with client id and client secret.

Metadata Service allows to register many services for particular Remote Environment. The Metadata Service API offers a great deal of flexibility as the user may decide how they want to expose their APIs.      

For a complete information on using Metadata Service API, please see [Managing registered services with Metadata API Guide](TODO) 

## API

Metadata Service API is described [here](TODO).


## Implementation Details

### Data storage

Service's data is stored in:
- [Remote Environment CRD](TODO)
- [Minio](TODO)
- Kubernetes secrets

Remote Environment CRD contains registered APIs and OAuth server urls. 

Minio storage contains:
- API specification
- Events catalog
- Documentation

### Kubernetes APIs usage 

Metadata Service interacts with Kubernetes APIs to perform the following tasks:
- modifying Remote Environment CRD instance
- creating secrets containing client id and client secret used to access OAuth secured APIs
- creating service used to access [Proxy Service](TODO) from lambda or service deployed on Kyma  

     