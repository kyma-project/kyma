---
title: What is Application Connectivity in Kyma?
---

Application Connectivity in Kyma is an area that: 

- Simplifies and secures the connection between external systems and Kyma
- Registers external events and APIs and simplifies the API usage
- Provides asynchronous communication with services and Functions deployed in Kyma through events
- Manages secure access to external systems
- Provides monitoring and tracing capabilities to facilitate operational aspects

Depending on your use case, Application Connectivity works in one of the two modes: 
- **Legacy mode** (default) - using components such as [Application Registry](ac-03-application-registry.md) and [Connector Service](docs/05-technical-reference/00-architecture/ac-02-connector-service.md)
- **Compass mode** - using [Runtime Agent](ra-01-runtime-agent-overview.md) and integration with [Compass](https://github.com/kyma-incubator/compass)