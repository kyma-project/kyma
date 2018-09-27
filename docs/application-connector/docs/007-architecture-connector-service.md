---
title: Connector Service
type: Architecture
---

## Overview

Connector Service is responsible for generating client certificates used to secure communication between Kyma and external systems.        
    
Creating a new client certificate is the first step of configuring external system represented by Remote Environment. Kyma stores root certificate and serves in this case as a Certificate Authority. In order to make validation possible the Connector Service returns complete certificate chain: created client certificate along with the root certificate. 

## API
    
There are two APIs exposed by the Connector Service:
- Internal API available inside Kyma cluster allowing to initiate certificate generation.
- External API exposed via Ingress allowing to finalize certificate generation. 

Both internal and external APIs are described [here](https://github.com/kyma-project/kyma/blob/master/docs/application-connector/docs/assets/connectorapi.yaml).

## Operation flow
 
Certificate generation is comprised of the following steps:
- Creating a token for particular Remote Environment using Connector Service Internal API.
- Getting information needed to create Certificate Signing Request using Connector Service External API.
- Creating a Certificate Signing Request.
- Creating a client certificate for specified CSR using Connector Service External API.   

The following diagram illustrates the operation flow in details:
![Client certificate generation operation flow](assets/002-automatic-configuration.png) 

For a complete information on using Connector Service APIs, please see [Obtaining the Certificate Guide](TODO). 

    
