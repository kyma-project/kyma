---
title: Connector Service
type: Details
---

## Overview
Connector Service is responsible for generating client certificates used to secure communication between Kyma and external systems.        
    

## Basic concepts 

Creating a new client certificate is the first step of configuring external system represented by Remote Environment. Kyma stores root certificate and serves in this case as a Certificate Authority. In order to make validation possible Connector Service returns complete certificate chain: created client certificate along with root certificate. 
    
There are two APIs exposed by Connector Service:
- Internal API available inside Kyma cluster allowing to initiate certificate generation
- External API exposed via Ingress allowing to finalize certificate generation 

Certificate generation is comprised of the following steps:
- creating a token for particular Remote Environment using Connector Service Internal API
- getting information needed to create Certificate Signing Request using Connector Service External API
- creating a Certificate Signing Request
- creating a client certificate for specified CSR using Connector Service External API   

For a complete information on using Connector Service, please see [Obtaining the Certificate Guide]() 

## API

Both internal and external APIs are described [here]().    
