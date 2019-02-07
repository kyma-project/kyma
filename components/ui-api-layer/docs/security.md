# Authentication

The application checks if Authorization token passed by user is signed by Dex. It is the standard way of performing Authentication in Kyma, which is usually done inside the Istio sidecars in components with Policy specified. In this application it is perfomed by middleware.

# Authorization

## Overview

For overview of authorization read [this document](/docs/security/docs/03-02-graphql.md).

## How to secure a query/mutation/subscription

All the queries/mutations/subscriptions are defined in [`schema.graphql`](../internal/gqlschema/schema.graphql) file in `components/ui-api-layer/internal/gqlschema/` directory. It is also the place where @HasAccess directive is defined. 

@HasAccess directive is used to secure the action or a field in a type. It is used as a middleware before the resolver code is executed. You can use the following query as an example on how to secure an action. You can find the detailed explanation in the table below.

```
limitRanges(namespace: String!): [LimitRange!]! @HasAccess(attributes: {resource: "limitranges", verb: "list", apiGroup: "", apiVersion: "v1", namespaceArg: "namespace", isChildResolver: false})
```

| Defined GraphQL action element | Description |
|:----------|:------|
| `limitRanges(namespace: String!)` |  Name of the action (in this case query) followed by a string that specifies the namespace of the queried resources. |
| `: [LimitRange!]!` | Defines the type of object returned by the query. In this case it's a list of LimitRanges resources. |
| `@HasAccess(attributes:` | Defines the GraphQL directive that secures the access to the resource. |
| `resource: "limitranges"` | Defines the type of secured Kubernetes resource, in this case all limitranges resources. |
| `verb: "list"` | Defines the secured interaction type. It is related to action type, which in this case is "query". |
| `apiGroup: ""` | Defines the apiGroup to which the user must have access to get the result of this query. In this case it is empty because limitRanges is the resource built into Kubernetes, not some Custom Resource created by us. |
| `apiVersion: "v1alpha1"` | Specifies the apiVersion of the query subject. |
| `namespaceArg: "namespace"` | Specifies the name of argument/field in parent object from which the resource namespace should be fetched |
| `isChildResolver: false` | (Optional) Specifies if directive is set on the field in a type. That means that that namespace should be fetched from parent object instead of the arguments, which are not available in "child". It is set to false by default. |