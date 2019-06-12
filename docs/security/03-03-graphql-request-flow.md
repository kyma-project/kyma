---
title: GraphQL request flow
type: Details
---

This diagram illustrates the request flow for the Console Backend Service which uses a custom [GraphQL](http://graphql.org/) implementation:

![GraphQL request flow](./assets/002-graphql-request-flow.svg)

1. The user sends a request with an ID token to the GraphQL application.
2. The GraphQL application validates the user token and extracts user data required to perform [Subject Access Review](https://kubernetes.io/docs/reference/access-authn-authz/authorization/#checking-api-access) (SAR).
3. The [Kubernetes API Server](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/) performs SAR.
4. Based on the results of SAR, the Kubernetes API Server informs the GraphQL application whether the user can perform the requested [GraphQL action](#details-graphql-available-graphql-actions).
5. Based on the information provided by the Kubernetes API Server, the GraphQL application returns an appropriate response to the user.

>**NOTE:** Read [this](#details-graphql) document to learn more about the custom GraphQL implementation in Kyma.
