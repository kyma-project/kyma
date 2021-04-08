# Authentication

The application checks if the Authorization token passed by the user is signed by Dex, which is the standard way of performing authentication in Kyma. In this application, it is performed by a middleware.

# Authorization

## Overview

The Console Backend Service uses a GraphQL implementation for authorization. Read [this](https://kyma-project.io/docs/master/components/security#details-graph-ql)) document for more details.

## How to secure a GraphQL action

All available GraphQL actions are defined in the [`schema.graphql`](../internal/gqlschema/schema.graphql) file. This file also contains the definition of the `@HasAccess` directive.

The `@HasAccess` directive is used to secure the action or a field in a type. It is used as a middleware before the resolver code is executed.

use the following query as an example on how to secure a GraphQL action:

```
limitRanges(namespace: String!): [LimitRange!]! @HasAccess(attributes: {resource: "limitranges", verb: "list", apiGroup: "", apiVersion: "v1", namespaceArg: "namespace", isChildResolver: false})
```

Reference this table for details on the elements that make up a defined and secured GraphQL action:

| Defined GraphQL action element | Description |
|----------|------|
| `limitRanges(namespace: String!)` |  Name of the action, which in this case is a query, followed by a string that specifies the Namespace of the queried resources. |
| `: [LimitRange!]!` | Defines the type of object returned by the query. In this case, it's a list of LimitRanges resources. |
| `@HasAccess(attributes:` | Defines the GraphQL directive that secures access to the resource. |
| `resource: "limitranges"` | Defines the type of secured Kubernetes resource, in this case all `limitranges` resources. |
| `verb: "list"` | Defines the secured interaction type. It is related to action type, which in this case is "query". Use the the table from [this paragraph](https://kyma-project.io/docs/master/components/security#details-graph-ql) to choose the right verb. |
| `apiGroup: ""` | Defines the apiGroup to which the user must have access to get the result of this query. In this case it is empty because limitRanges is the resource built into Kubernetes, not some Custom Resource created by us. |
| `apiVersion: "v1alpha1"` | Specifies the apiVersion of the query subject. |
| `namespaceArg: "namespace"` | Specifies the name of the argument or field in the parent object from which the resource namespace is fetched. |
| `isChildResolver: false` | Must be "true" for fields nested in types which have to be secured. Determines if the `namespace` argument should be fetched from the parent object. By default, it is set to `false`. |

If the directive is set on a field nested in a type, the value returned from a query returning that type depends on the type modifier set on the field. If the field is set as [`Non-Null`](https://graphql.org/learn/schema/#lists-and-non-null), the error is returned for users who do not have rights to access the child resource. If the field is not set as `Non-Null`, only the child field will be returned as null.
