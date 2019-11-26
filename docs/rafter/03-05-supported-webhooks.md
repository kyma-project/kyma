---
title: Supported webhooks
type: Details
---

## Types

Rafter supports the following types of webhooks:

- **Mutation webhook** modifies fetched assets before the Asset Controller uploads them into the bucket. For example, this can mean asset rewriting through the `regex` operation or `key-value`, or the modification in the JSON specification. The mutation webhook service must return modified files to the Asset Controller.

- **Validation webhook** validates fetched assets before the Asset Controller uploads them into the bucket. It can be a list of several different validation webhooks that process assets even if one of them fails. It can refer either to the validation of a specific file against a specification or to the security validation. The validation webhook service must return the validation status when the validation completes.

- **Metadata webhook** allows you to extract metadata from assets and inserts it under the `status.assetRef.files.metadata` field in the (Cluster)Asset CR. For example, the Asset Metadata Service which is the metadata webhook implementation in Kyma, extracts front matter metadata from `.md` files and returns the status with such information as `title` and `type`.

## Service specification requirements

If you create a specific mutation, validation, or metadata service for the available webhooks and you want Rafter to properly communicate with it, you must ensure that the API exposed by the given service meets the API contract requirements. These criteria differ depending on the webhook type:

>**NOTE:** Services are described in the order in which Rafter processes them.

- **mutation service** must expose endpoints that:

  - accept **parameters** and **content** properties.
  - return the `200` response with new file content.
  - return the `304` response informing that the file content was not modified.

- **validation service** must expose endpoints that:

  - contain **parameters** and **content** properties.
  - return the `200` response confirming that validation succeeded.
  - return the `422` response informing why validation failed.

- **metadata service** must expose endpoints that:

  - pass file data in the `"object": "string"` format in the request body, where **object** stands for the file name and **string** is the file content.
  - return the `200` response with extracted metadata.

See the example of an API specification with the `/convert`, `/validate`, and `/extract` endpoints [here](./assets/example-openapi-service.yaml).
