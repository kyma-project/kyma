---
title: Resources
type: Details
---

This document includes an overview of resources that the Kyma Application Connector provides.

* **Client certificate** is the X509 certificate which identified the connected system and ensures the secured connection.
* **Event catalog** documents all events which external system can send to the Kyma.
* **Connector service** establishes a secure connection between external system and Kyma. It provides the client certificate to the connected system.
* **Metadata service** manage the APIs and Event catalog registration within Kyma.
* **Event service** delivers events to Kyma Eventbus.
* **Proxy service** proxing call from Kyma to connected systems' APIs.
* **Remote Environment** represents connected system in Kyma. There is always one to one relation between the Remote Environment and connected system.
