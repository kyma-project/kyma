---
title: CMS AsyncAPI Service
type: Details
---

CMS AsyncAPI Service is an HTTP server used in Kyma to process AsyncAPI specifications. It only accepts `multipart/form-data` forms and contains two endpoints:

- `/validate` that validates the AsyncAPI specification against the AsyncAPI schema in version 2.0.0.

- `/convert` that converts the AsyncAPI files version and format. The service uses the [AsyncAPI Converter](https://github.com/asyncapi/converter-go) to change the AsyncAPI specifications from older versions to version 2.0.0, and convert any `yaml` input files to the `json` format that is required to render the specifications in the Console UI.

See the [this](https://github.com/kyma-project/kyma/blob/master/components/cms-services/cmd/asyncapi/openapi.yaml) file for the full OpenAPI specification of the service. To preview and test the API service, use the [Swagger Editor](https://editor.swagger.io/).
