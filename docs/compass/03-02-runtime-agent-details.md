---
title: Runtime Agent
type: Details
---

Runtime Agent is a Kyma component that connects to Compass. 

The main responsibilities of the component are:
- [Establishing a trusted connection](08-07-establish-secure-connection-with-compass.md) between the Kyma Runtime and Compass
- [Renewing a trusted connection](08-08-maintain-secure-connection-with-compass.md) between the Kyma Runtime and Compass
- Synchronizing with the Director by fetching new Applications from the Director and creating them in the Runtime, and removing from the Runtime Applications that no longer exist in the Director.