---
title: Managing Lambdas
type: Details
---

Kubernetes provides Kyma with labels that allow you to arrange lambda functions and group them. Labeling also makes it possible to filter lambdas functions. This functionality is particularly useful when a developer needs to manage a large set of lambda functions.

Behind the scenes, labeling takes place in the form of key value pairs. Here is an example of code that enhances a function:

```
"labels": {
  "key1" : "value1",
  "key2" : "value2"
}
```

For more details on labels and selectors, visit the [Kubernetes website](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/).