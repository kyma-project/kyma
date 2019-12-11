# AsyncAPI Service

## Overview

The AsyncAPI Service is an HTTP server used to process AsyncAPI specifications. It contains the `/validate` and `/convert` HTTP endpoints which accept `multipart/form-data` forms:

- The `/validate` endpoint validates the AsyncAPI specification against the AsyncAPI schema in version 2.0.0., using the [AsyncAPI Parser](https://github.com/asyncapi/parser).
- The `/convert` endpoint converts the version and format of the AsyncAPI files.

To learn more about the AsyncAPI Service, go to the [Rafter repository](https://github.com/kyma-project/rafter/tree/master/cmd/extension/asyncapi).
