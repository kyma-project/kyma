---
title: Application Gateway
---

Central Application Gateway is an intermediary component between a Function or a service and an external API.  
This component is deployed in ```kyma-system``` Namespace. 
Its role is to proxy HTTP requests from user's Functions to all External Systems registered in Kyma as Applications. 

Central Application Gateway can call services which are not secured, or are secured with:

- [Basic Authentication](https://tools.ietf.org/html/rfc7617)
- OAuth
- Client certificates

Additionally, Central Application Gateway supports cross-site request forgery (CSRF) tokens as an optional layer of API protection.