# Application Registry

## Overview
Application Registry is a global component responsible for managing metadata of remote APIs.

## Configuration
This service has the following parameters:

- **proxyPort** - This port is used for services created for the Proxy Service. The default port is `8080`.
- **externalAPIPort** - This port exposes the metadata API to an external system. The default port is `8081`.
- **minioURL** - The URL of a Minio service which stores specifications and documentation.
- **namespace** - The Namespace to which you deploy the Proxy Service. The default Namespace is `kyma-integration`.
- **requestTimeout** - A time-out for requests sent through the Proxy Service. Provide it in seconds. The default time-out is `10`.
- **requestLogging** - The flag to enable logging of incoming requests. The default value is `false`.
- **specRequestTimeout** - A time-out for requests fetching specifications provided by the user. It is provided in seconds. The default time-out is `5`.
- **assetstoreRequestTimeout** - A time-out for requests fetching specifications from the Asset Store Service. It is provided in seconds. The default time-out is `5`.
- **detailedErrorResponse** - A flag for showing detailed internal error messages in response bodies. The default value is `false` and all internal server error messages are shortened to `Internal error`, while all other error messages are shown as usual.
