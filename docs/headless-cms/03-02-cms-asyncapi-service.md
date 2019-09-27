---
title: CMS AsyncAPI Service
type: Details
---

CMS AsyncAPI Service is an HTTP server enabled by default in Kyma to process AsyncAPI specifications. It only accepts `multipart/form-data` forms and contains two endpoints:

- `/validate` that validates the AsyncAPI specification against the AsyncAPI schema in version 2.0.0. CMS AsyncAPI Service uses the [AsyncAPI Parser](https://github.com/asyncapi/parser) for this purpose.

- `/convert` that converts the version and format of the AsyncAPI files. The service uses the [AsyncAPI Converter](https://github.com/asyncapi/converter-go) to change the AsyncAPI specifications from older versions to version 2.0.0, and convert any `yaml` input files to the `json` format that is required to render the specifications in the Console UI.

See [this](https://github.com/kyma-project/kyma/blob/master/components/cms-services/cmd/asyncapi/openapi.yaml) file for the full OpenAPI specification of the service. To preview and test the API service, use the [Swagger Editor](https://editor.swagger.io/).

>**NOTE:** To learn how you can configure the service with an override, see [this](#configuration-cms-asyncapi-service-sub-chart) document.
