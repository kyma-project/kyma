---
title: Overview
---

>**CAUTION:** This implementation will soon replace Kyma Event Bus. Consider it as experimental as it is still under development.

Knative Eventing Mesh allows you to easily integrate external applications with Kyma. Under the hood, Knative Eventing Mesh implements [Knative Eventing](https://knative.dev/docs/eventing/) to ensure Kyma receives business events from external sources and is able to trigger business flows using lambda functions or services. 

Knative Eventing Mesh implementation relies on [Knative Broker and Trigger](https://knative.dev/docs/eventing/broker-trigger/) custom resources that define the event delivery process. 
Knative Broker receives events from external solutions and dispatches them to Subscribers, such as lambda functions.
To make sure certain Subscribers receive exactly the events they want, Knative Trigger defines filters that use event attributes, such as version or type, to pick up specific events from Knative Broker. 

As a result, the process of event publishing and consumption runs smoother, thus significantly improving the overall performance. 

