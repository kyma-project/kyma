---
title: Application Gateway
---

<!-- TODO 2: add mention that Gateway is a central component -->

Application Gateway is an intermediary component between a Function or a service and an external API.  

Application Gateway can call services which are not secured, or are secured with:

- [Basic Authentication](https://tools.ietf.org/html/rfc7617)
- OAuth
- Client certificates

Additionally, Application Gateway supports cross-site request forgery (CSRF) tokens as an optional layer of API protection.