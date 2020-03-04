---
title: Overview
---

>**CAUTION:** This implementation will soon replace Kyma Event Bus. Consider it as experimental as it is still under development.

Knative Eventing Mesh allows you to easily integrate external applications with Kyma. To achieve successful integration, the Eventing Mesh uses [Knative Eventing](https://knative.dev/docs/eventing/) to ensure that Kyma receives business events from external sources and triggers business flows using lambda functions or services. 

Knative Eventing Mesh implementation relies on [Knative Broker and Trigger](https://knative.dev/docs/eventing/broker-trigger/) custom resources that define the event delivery process. 
The Broker receives events from external solutions and dispatches them to Subscribers, such as lambda functions. To make sure certain Subscribers receive exactly those events they want, Trigger definition specifies filters that use event attirbutes, such as version or type, to filter events.

As a result, the process of event publishing and consumption runs smoother, thus significantly improving the overall performance. 

