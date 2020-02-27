---
title: Overview
---


>**CAUTION:** This implementation will soon replace Event Bus. Consider this implementation as experimental as it is still under development.


Knative Eventing Mesh allows you to integrate various external solutions with Kyma. To achieve successful integration, Eventing Mesh uses [Knative Eventing](https://knative.dev/docs/eventing/) to make sure Kyma receives business Events from different solutions and is able to enrich them, and trigger business flows using lambdas or services defined in Kyma. 

Eventing Mesh impementation relies on [Knative Broker and Trigger](https://knative.dev/docs/eventing/broker-trigger/) Custom Resources, which define the logic behind event processing. The Broker receives events from solutions and forwards them to Subscribers, such as lambda functions, based on the defined filters.
This way, events can come from different Senders and Subscribers receive exactly those events they want to. 
As a result, process of event publishing and consumption runs smoother, thus significantly improving the overall performance. 

