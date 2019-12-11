# Upload Service

## Overview

The Upload Service is an HTTP server used for hosting static files in [MinIO](https://min.io/). It contains a simple HTTP endpoint which accepts `multipart/form-data` forms. It uploads files to dedicated private and public system buckets that the service creates in MinIO, instead of Rafter. This service is particularly helpful if you do not have your own storage place from which Rafter could fetch assets. You can also use this service for development purposes to host files temporarily, without the need to rely on external providers.

To learn more about the Upload Service, go to the [Rafter repository](https://github.com/kyma-project/rafter/tree/master/cmd/uploader).
