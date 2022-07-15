---
title: What is Eventing in Kyma?
---

Eventing in Kyma is an area that:

- allows users to easily connect external applications with Kyma.
- implements [NATS JetStream](https://docs.nats.io/) internally within the cluster, to ensure Kyma receives business events from external sources.
- triggers business flows using workloads such as Functions or services.
- enables users to use events to implement asynchronous flows within Kyma, as the source and the sink can be inside Kyma.
- simplifies sending events using HTTP POST requests.
