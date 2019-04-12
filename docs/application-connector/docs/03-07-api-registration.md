---
title: API registration in Application Registry
type: Details
---

### Supported formats

The Application Registry supports API Specification in the following formats:
- OpenAPI 2.0
- OData 2.0

### Registration methods

Application Registry supports two ways of passing API spec:
- in JSON format
- by `SpecificationUrl`

>**NOTE:** Specification passed directly as a JSON has a higher priority than `SpecificationUrl`, which in such case will be ignored

For OpenAPI format both methods are supported.
OData APIs can only be registered with `SpecificationUrl`.
