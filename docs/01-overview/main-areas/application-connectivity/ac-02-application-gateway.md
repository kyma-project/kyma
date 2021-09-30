---
title: Application Gateway
---

Application Gateway is an intermediary component between a Function or a service and an external API.  
It [proxies the requests](../../../05-technical-reference/00-architecture/ac-03-application-gateway.md) from Functions and services in Kyma to external APIs based on the configuration stored in Secrets.
 
Application Gateway can call services which are not secured, or are secured with:

- [Basic Authentication](https://tools.ietf.org/html/rfc7617)
- OAuth
- Client certificates

Additionally, Application Gateway supports cross-site request forgery (CSRF) tokens as an optional layer of API protection.