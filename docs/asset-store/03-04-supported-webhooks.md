---
title: Supported webhooks
type: Details
---

## Types

Asset Store supports the following types of webhooks:

- **Mutation webhook** mutates fetched assets. For example, this can mean asset rewriting through the `regex` operation or `keyvalue`, or the modification in the JSON specification. The mutation webhook returns modified files instead of information on the status.

- **Validation webhook** validates fetched assets before the Asset Controller uploads them into the bucket. It can be a list of several different validation webhooks and all of them should be processed even if one fails. It can refer either to the validation of a specific file against a specification or to the security validation. The validation webhook returns the validation status when the validation completes.

- **Metadata webhook** allows you to extract metadata from assets and inserts it under the `status.assetRef.files.metadata` field in the (Cluster)Asset CR. For example, the Asset Metadata Service which is the metadata webhook implementation in Kyma, extracts front matter metadata from `.md` files and returns the status with such information as `title` and `type`. The metadata is later used to render documentation in the Console UI.   

## Service specification requirements

If you create a specific mutation, validation, or metadata service for the available webhooks and you want the Asset Store to properly communicate with it, you must ensure that the API exposed by the given service meets certain format requirements. These criteria differ depending on the service type:

> **NOTE:** Services are described in the order in which the Asset Store processes them.

- **mutation service** must expose endpoints that:

  - accept **parameters** and **content** properties.
  - return the `200` response with new file content.
  - return the `304` response.

See the example of the validation service with the `/convert` endpoint [here](./assets/mutation-service.yaml). To preview and test it, use the [Swagger Editor](https://editor.swagger.io).

- **validation service** must expose endpoints that:

  - contain **parameters** and **content** properties.
  - return the `200` response with new file content.

See the example of the validation service with the `/validate` endpoint [here](./assets/validation-service.yaml). To preview and test it, use the [Swagger Editor](https://editor.swagger.io).

- **metadata service** must expose endpoints that:

  - pass file data in the in `"object": "string"` format in the request body, where **object** stands for the file name and **string** is the file content.
  - return the `200` response with new file content.

See the example of the validation service with the `/extract` endpoint [here](./assets/metadata-service.yaml). To preview and test it, use the [Swagger Editor](https://editor.swagger.io).
