---
title: Architecture
type: Architecture
---

The Application Connector consists of the following components:

* **Ingress-Gateway controller** validating certificates and exposing multiple Application Connectors to the external world.
* **Gateway** proxying calls to the registered solution.
* **Remote Environment CRD instance** storing a solution's metadata.
* **Remote Environtment Controller** provisioning and de-provisioning necessary deployments necessary deployments for the created Remote Environments.
* **Minio bucket** storing API specifications, Event Catalogs, and documentations.

To connect a new external solution, you must install and set up a new Remote Environment. Every external solution connected to Kyma is a separate Remote Environment with a dedicated Gateway Service and a dedicated Event Service. See the **Set up a Remote Environment on local Kyma installation** getting started guide to learn how to connect an external solution to Kyma.

![Architecture Diagram](assets/001-application-connector.png)
