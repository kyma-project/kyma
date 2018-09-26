---
title: Resources
type: Details
---

This document includes an overview of resources that the Kyma Application Connector provides.

* **Client certificate** is the X509 certificate which identified the connected system and ensures the secured connection.
* **Event catalog** is the documentation of all events which external system can send to the Kyma.
* **Connector service** is responsible for establishing a secure connection between external system and Kyma. It provides the client certificate to the connected system.
* **Metadata service** is responsible for managing the APIs and Event catalog registration within Kyma.
* **Event service** is responsible for delivering events to Kyma Eventbus.
* **Proxy service** is reponsible for proxing call from Kyma to connected systems' APIs.
* **Remote Environment** is representing the connected system in Kyma. There is always one to one relation between the Remote Environment and connected system.
