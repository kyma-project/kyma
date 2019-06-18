---
title: Blacklisted services in the API Controller
type: Details
---

The API Controller uses a blacklist of services for which it doesn't create either a Virtual Service or Authentication Policies. As a result, these services cannot be exposed. Every time a user creates a new Api custom resource (CR) for a service, the API Controller checks the service name specified in the CR against the blacklist.

The blacklist works as a security measure and prevents users from exposing vital internal services of Kubernetes, Istio, and API Server Proxy.   
