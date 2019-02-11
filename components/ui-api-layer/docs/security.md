# authentication

the application checks if the authorization token passed by the user is signed by dex, which is the standard way of performing authentication in kyma. in this application, it is performed by a middleware.

# authorization

## overview

the ui api layer uses a graphql implementation for authorization. read [this](https://kyma-project.io/docs/master/components/security#details-graphql) document for more details.

## how to secure a graphql action

all available graphql actions are defined in the [`schema.graphql`](../internal/gqlschema/schema.graphql) file. this file also contains the definition of the `@hasaccess` directive.

the `@hasaccess` directive is used to secure the action or a field in a type. it is used as a middleware before the resolver code is executed. 

use the following query as an example on how to secure a graphql action:
```
limitranges(namespace: string!): [limitrange!]! @hasaccess(attributes: {resource: "limitranges", verb: "list", apigroup: "", apiversion: "v1", namespacearg: "namespace", ischildresolver: false})
```

reference this table for details on the elements that make up a defined and secured graphql action:

| defined graphql action element | description |
|:----------|:------|
| `limitranges(namespace: string!)` |  name of the action, which in this case is a query, followed by a string that specifies the namespace of the queried resources. |
| `: [limitrange!]!` | defines the type of object returned by the query. in this case, it's a list of limitranges resources. |
| `@hasaccess(attributes:` | defines the graphql directive that secures access to the resource. |
| `resource: "limitranges"` | defines the type of secured kubernetes resource, in this case all `limitranges` resources. |
| `verb: "list"` | defines the secured interaction type. it is related to action type, which in this case is "query". use the the table from [this paragraph](https://kyma-project.io/docs/master/components/security#details-graphql) to choose the right verb. |
| `apigroup: ""` | defines the apigroup to which the user must have access to get the result of this query. in this case it is empty because limitranges is the resource built into kubernetes, not some custom resource created by us. |
| `apiversion: "v1alpha1"` | specifies the apiversion of the query subject. |
| `namespacearg: "namespace"` | specifies the name of the argument or field in the parent object from which the resource namespace is fetched. |
| `ischildresolver: false` | required to be "true" for fields nested in types which have to be secured. determines if the namespace argument should be fetched from the parent object. by default it is set to `false`. |

If the directive is set on a field nested in a type, the value returned from a query returning that type depends on the type modifier set on the field. If the field is set as Non-Null, as described [here](https://graphql.github.io/learn/schema/#lists-and-non-null), the error will be returned in case the user has no rights to access the child resource. If the field is not set as Non-Null, only the child field will be returned as null.