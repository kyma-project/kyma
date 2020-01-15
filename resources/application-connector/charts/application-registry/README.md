# Application Registry

## Overview
Application Registry is a global component responsible for managing metadata of remote APIs.

## Configuration
This service has the following parameters:

- **proxyPort** is the port used for services created for the Proxy Service. The default port is `8080`.
- **externalAPIPort** is the port that exposes the metadata API to an external system. The default port is `8081`.
- **minioURL** is the URL of a MinIO service which stores specifications and documentation.
- **namespace** is the Namespace to which you deploy the Proxy Service. The default Namespace is `kyma-integration`.
- **requestTimeout** is the time-out for requests sent through the Proxy Service. Provide it in seconds. The default time-out is `10`.
- **requestLogging** is the flag to enable logging of incoming requests. The default value is `false`.
- **specRequestTimeout** is the time-out for requests fetching specifications provided by the user. It is provided in seconds. The default time-out is `20`.
- **rafterRequestTimeout** is the time-out for requests fetching specifications from Rafter. It is provided in seconds. The default time-out is `20`.
- **detailedErrorResponse** - is the flag for showing detailed internal error messages in response bodies. The default value is `false` and all internal server error messages are shortened to `Internal error`, while all other error messages are shown as usual.
- **insecureAssetDownload** - is the flag for skipping certificate verification for asset download.
- **insecureSpecDownload** - is the flag for skipping certificate verification for API specification download.
