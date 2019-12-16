---
title: AsyncAPI Service
type: Details
---

AsyncAPI Service is an HTTP server enabled by default in Kyma to process AsyncAPI specifications. It only accepts `multipart/form-data` forms and contains two endpoints:

- `/validate` that validates the AsyncAPI specification against the AsyncAPI schema in version 2.0.0. AsyncAPI Service uses the [AsyncAPI Parser](https://github.com/asyncapi/parser) for this purpose.

- `/convert` that converts the version and format of the AsyncAPI files. The service uses the [AsyncAPI Converter](https://github.com/asyncapi/converter-go) to change the AsyncAPI specifications from older versions to version 2.0.0, and convert any YAML input files to the JSON format that is required to render the specifications in the Console UI.

See [this](./assets/asyncapi-service-openapi.yaml) file for the full OpenAPI specification of the service.
