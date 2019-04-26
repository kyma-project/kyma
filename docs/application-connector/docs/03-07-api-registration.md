---
title: API registration in the Application Registry
type: Details
---

The Application Registry supports the following formats of the API specification:
- OpenAPI 2.0
- OData XML 2.0, 3.0 and 4.0

You can pass the API specification in two ways:
- JSON format
- `SpecificationUrl`

>**NOTE:** Specification passed directly as a JSON has a higher priority than `SpecificationUrl`.  If you use these two methods at once, `SpecificationUrl` is ignored.

For the OpenAPI format, both methods are supported.
You can register OData APIs only with `SpecificationUrl`.
