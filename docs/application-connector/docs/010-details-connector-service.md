---
title: Connector Service
type: Details
---

The Connector Service generates client certificates which secure the communication between Kyma and the connected external solutions.        

Generating a new client certificate is the first step in the process of configuring a Remote Environment (RE). Kyma stores the root certificate and serves as the Certificate Authority when you configure a new RE. When you generate a new client certificate, the Connector Service returns it along with the root certificate to allow validation.  

This diagram illustrates the client certificate generation flow in details:
![Client certificate generation operation flow](assets/002-automatic-configuration.png)
