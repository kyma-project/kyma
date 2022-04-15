---
title: What is Application Connectivity in Kyma?
---

Application Connectivity in Kyma is an area that: 

- Simplifies and secures the connection between external systems and Kyma
<!-- TODO: is this one still true?
- Registers external events and APIs and simplifies the API usage
-->
<!-- TODO: is this one still true?
- Provides asynchronous communication with services and Functions deployed in Kyma through events
-->
- Manages secure access to external systems
- Provides monitoring and tracing capabilities to facilitate operational aspects

Depending on your use case, Application Connectivity works in one of the two modes: 
- **Legacy mode** (default) - a standalone mode where Kyma is not connected to Compass
- **Compass mode** - using [Runtime Agent](ra-01-runtime-agent-overview.md) and integration with [Compass](https://github.com/kyma-incubator/compass)