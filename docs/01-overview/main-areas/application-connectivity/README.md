---
title: What is Application Connectivity in Kyma?
---

Application Connectivity in Kyma is an area that: 

- Simplifies and secures the connection between external systems and Kyma
- Stores and handles the metadata of external systems
- Provides certificate handling for the [Eventing](../eventing/README.md) flow in the Compass scenario (mode)
- Manages secure access to external systems
- Provides monitoring and tracing capabilities to facilitate operational aspects

Depending on your use case, Application Connectivity works in one of the two modes: 
- **Standalone mode** (default) - a standalone mode where Kyma is not connected to [Compass](https://github.com/kyma-incubator/compass)
- **Compass mode** - using [Runtime Agent](ra-01-runtime-agent-overview.md) and integration with [Compass](https://github.com/kyma-incubator/compass) to automate connection and registration of services using mTLS certificates