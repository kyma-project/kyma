---
title: Overview
---

Lambdas or lambda functions are small functions that run in Kyma in a cost-efficient and scalable way using JavaScript in Node.js. As the following diagram shows, these functions enable the linking of a wide range of functionalities using Kyma.

![Kyma connected to other products through lambda functions](./assets/kyma_connected.png)


This is an example lambda function:

```
module.exports = {
  foo: function (event, context) {
    return 'hello world!';
  }
}
```

The use of lambdas in Kyma addresses several scenarios:  

 * Create and manage lambda functions
 * Trigger functions based on business Events
 * Expose functions through HTTP
 * Consume services
 * Provide customers with customized features
 * Version lambda functions
 * Chain multiple functions

Lambda functions best serve integration purposes due to their ease of use. Lambda is a quick and ideal solution when the goal is to combine functionalities which are tightly coupled. And, in the context of Kyma, they provide integration with the Event system and Customer Engagement and Commerce tools. Lambda functions are not well-suited to building an application from scratch.

The Serverless implementation of Kyma is based on [Kubeless](https://github.com/kubeless/kubeless).
