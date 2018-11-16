### Metadata service

## Overview
Metadata service is a global component responsible for managing metadata of remote APIs.

## Configuration
This service has the following parameters:

- **proxyPort** - This port is used for services created for the Gateway proxy. The default port is `8080`.
- **externalAPIPort** - This port exposes the metadata API to an external system. The default port is `8081`.
- **minioURL** - The URL of a Minio service which stores specifications and documentation.
- **namespace** - The Namespace to which you deploy the Gateway. The default Namespace is `kyma-integration`.
- **requestTimeout** - A time-out for requests sent through the Gateway. Provide it in seconds. The default time-out is `1`.
- **requestLogging** - The flag to enable logging of incoming requests. The default value is `false`.