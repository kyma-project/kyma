---
title: Overview
---

>**CAUTION:** This implementation will soon replace the Event Bus. Consider it as experimental as it is still under development.

Knative Eventing Mesh allows you to easily integrate external applications with Kyma. Under the hood, the Eventing Mesh implements [Knative Eventing](https://knative.dev/docs/eventing/) to ensure Kyma receives business events from external sources and is able to trigger business flows using lambda functions or services. 

Knative Eventing Mesh implementation relies on Knative's [Broker and Trigger](https://knative.dev/docs/eventing/broker-trigger/) custom resources that define the event delivery process. 
The Broker receives events from external solutions and dispatches them to subscribers, such as lambda functions.
To make sure certain subscribers receive exactly the events they want, the Trigger defines filters that use event attributes, such as version or type, to pick up specific events from the Broker. 


