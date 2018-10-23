# Application Connector

## Overview

The Application Connector connects an external solution to Kyma.

## Details

The Application Connector Helm chart contains all the global components:
- Metadata service
- Connector service

### Metadata service

Metadata service is a global component responsible for managing metadata of remote APIs.

This service has the following parameters:

- **proxyPort** - This port is used for services created for the Gateway proxy. The default port is `8080`.
- **externalAPIPort** - This port exposes the metadata API to an external system. The default port is `8081`.
- **minioURL** - The URL of a Minio service which stores specifications and documentation.
- **namespace** - The Namespace to which you deploy the Gateway. The default Namespace is `kyma-integration`.
- **requestTimeout** - A time-out for requests sent through the Gateway. Provide it in seconds. The default time-out is `1`.
- **requestLogging** - The flag to enable logging of incoming requests. The default value is `false`.

### Connector service

Connector service is a global component responsible for automatic certificate configuration for external systems.

The Connector Service has the following parameters:
- **appName** - This is the name of the application used by Kubernetes Deployments and services.
- **externalAPIPort** - This port exposes the Connector Service API to an external system.
- **internalAPIPort** - This port exposes the Connector Service within the Kubernetes cluster.
- **namespace** - Namespace where the Connector Service is deployed.
- **tokenLength** - Length of registration tokens.
- **tokenExpirationMinutes** - Time after which tokens expire and are no longer valid.
- **domainName** - Domain name of the cluster, used for URL generating.
- **certificateServiceHost** - Host at which this service is accessible, used for URL generating.

### Installation

The Application Connector is a part of the Kyma core and it installs automatically.
