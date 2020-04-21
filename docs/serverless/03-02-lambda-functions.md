---
title: Lambda functions
type: Details
---

Lambda functions best serve integration purposes due to their ease of use. Lambda is a quick solution when the goal is to combine functionalities which are tightly coupled. In Kyma, lambdas run in a cost-efficient and scalable way using JavaScript in Node.js. 

This is an example lambda function:

```
def myfunction (event, context):
  print event
  return event['data']
```

In Kyma, you can address the following scenarios: 

 * Create and manage lambda functions.
 * Trigger functions based on business Events.
 * Expose functions through HTTP.
 * Consume services.
 * Provide customers with customized features.
 * Version lambda functions.
 * Chain multiple functions.

Kubernetes provides Kyma with labels that allow you to arrange lambda functions and group them. Labeling also makes it possible to filter lambdas functions. This functionality is particularly useful when a developer needs to manage a large set of lambda functions.

Behind the scenes, labeling takes place in the form of key-value pairs. Here is the example of code that enhances a function:

```
"labels": {
  "key1" : "value1",
  "key2" : "value2"
}
```

For more details on labels and selectors, visit the [Kubernetes website](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).
