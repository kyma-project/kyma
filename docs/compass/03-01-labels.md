---
title: Labels
type: Details
---

A label is a key-value pair that you can add to every top-level entity, such as an Application or a Runtime. Labels allow you to enrich Applications and Runtimes with additional information.

Use GraphQL mutations to set labels on your Applications and Runtimes. For example, to set the `test:val` label on your Application, use:

```graphql
mutation {
  result: setApplicationLabel(
    applicationID: "{YOUR_APPLICATION_ID}"
    key: "test"
    value: "val"
  ) {
    key
    value
  }
}
```

## LabelDefinitions

For every label, you can create a related LabelDefinition to set validation rules for values using a JSON schema. It allows you to search for all Applications and Runtimes label keys used in a given tenant. LabelDefinitions are optional, but they are created automatically if the user adds labels for which LabelDefinitions do not exist. You can manage LabelDefinitions using the following mutations and queries:

```graphql
type Query {
    labelDefinitions: [LabelDefinition!]!
    labelDefinition(key: String!): LabelDefinition
}

type Mutation {
    createLabelDefinition(in: LabelDefinitionInput!): LabelDefinition!
    updateLabelDefinition(in: LabelDefinitionInput!): LabelDefinition!
    deleteLabelDefinition(key: String!, deleteRelatedLabels: Boolean=false): LabelDefinition!
}
```

> **TIP:** For all the GraphQL query and mutation examples that you can use, go to [this](https://github.com/kyma-incubator/compass/tree/master/components/director/examples) directory.

LabelDefinition key has to be unique for a given tenant. You can provide only one value for a given label. However, this label can contain many elements, depending on the LabelDefinition schema. In the following example, the type of the label value is `Any`, which means that the value can be of any type, such as `JSON`, `string`, `int` etc.:

```graphql
setApplicationLabel(applicationID: ID!, key: String!, value: Any!): Label!
```

### Define and set LabelDefinitions

See the example of how you can create and set a LabelDefinition:

```
createLabelDefinition(in: {
  key:"supportedLanguages",
  schema:"{
              "type": "array",
              "items": {
                  "type": "string",
                  "enum": ["Go", "Java", "C#"]
              }
          }"
}) {...}


setApplicationLabel(applicationID: "123", key: "supportedLanguages", value:["Go"]) {...}

```

### Edit LabelDefinitions

You can edit LabelDefinitions using a mutation. When editing a LabelDefinition, make sure that all labels are compatible with the new definition. Otherwise, the mutation is rejected with a clear message that there are Applications or Runtimes that have an invalid label according to the new LabelDefinition. In such a case, you must either:
- Remove incompatible labels from specific Applications and Runtimes, or
- Remove the old LabelDefinition from all labels using cascading deletion.

For example, let's assume we have the following LabelDefinition:

```graphql
 key:"supportedLanguages",
  schema:{
              "type": "array",
              "items": {
                  "type": "string",
                  "enum": ["Go", "Java", "C#"]
              }
          }
```

If you want to add a new language to the list, provide such a mutation:

```
updateLabelDefinition(in: {
                        key:"supportedLanguages",
                        schema:{
                                    "type": "array",
                                    "items": {
                                        "type": "string",
                                        "enum": ["Go", "Java", "C#","ABAP"]
                                    }
                                }
                      }) {...}
```

### Remove LabelDefinitions

Use this mutation to remove a LabelDefinition:

```graphql
deleteLabelDefinition(key: String!, deleteRelatedLabels: Boolean=false): LabelDefinition

```
This mutation allows you to remove only definitions that are not used. If you want to delete a LabelDefinition with all its values, set the **deleteRelatedLabels** parameter to `true`.

## LabelFilters

You can define a LabelFilter to list the top-level entities according to their labels. You can search for Applications and Runtimes by label keys or by label keys and their values. To search for a given Application or Runtime, use this query:

```graphql
 applications(filter: [LabelFilter!], first: Int = 100, after: PageCursor):  ApplicationPage!
```

To search for all objects with a given label ignoring their values, use:

```graphql
query {
  applications(filter:[{key:"scenarios"
    }]) {
    data {
      name
      labels
    }
    totalCount
  }
}
```

To filter objects by their key and string value, use this query:

```graphql
runtimes(filter: { key: "{KEY}" query: "\"{VALUE}\"" })
```

You can also search for objects by their key and array values. In the **query** field, use only the limited SQL/JSON path expressions. The supported syntax is `$[*] ? (@ == "{VALUE}" )`. For example, to filter all objects assigned to the `default` scenario, run:

```graphql
query {
  applications(filter:[{key:"scenarios",
    query:"$[*] ? (@ == \"DEFAULT\")"}]) {
    data {
      name
      labels
    }
    totalCount
  }
}
```

## Scenarios label

Every Application is labeled with the special **Scenarios** label which automatically has the `default` value assigned. As every Application has to be assigned to at least one scenario, if no scenarios are explicitly specified, the `default` scenario is used.

When you create a new tenant, the **Scenarios** LabelDefinition is created. It defines a list of possible values that can be used for the **Scenarios** label. Every time you create or modify an Application, there is a step that ensures that **Scenarios** label exists. You can add or remove values from the **Scenarios** LabelDefinition list, but neither the `default` value, nor the **Scenarios** label can be removed.
