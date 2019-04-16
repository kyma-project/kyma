---
title: API registration in the Application Registry
type: Details
---

The Application Registry supports the following formats of the API specification:
- OpenAPI 2.0
- OData XML 2.0, 3.0 and 4.0

You can pass the API specification in two ways:
- JSON format
- by `SpecificationUrl`

>**NOTE:** Specification passed directly as a JSON has a higher priority than `SpecificationUrl`, which in such case will be ignored.

For OpenAPI format both methods are supported.
OData APIs can only be registered with `SpecificationUrl`.
