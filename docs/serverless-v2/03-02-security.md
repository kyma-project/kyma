---
title: Security
type: Details
---

To eliminate potential security risks when using lambdas, bear in mind these few facts:

- Kyma does not run any security scans against lambdas and their images. Before you store any sensitive data in lambdas, consider the potential risk of data leakage.

- By default, JSON Web Tokens (JWTs) issued by Dex do not provide the **scope** parameter for lambdas. This means that if you expose your lambda and secure it with a JWT, you can use the token to validate access to all lambdas within the cluster.

- Kyma does not define any authorization policies that would restrict lambdas' access to other resources within the Namespace. If you deploy a lambda in a given Namespace, it can freely access all events and APIs of services within this Namespace.

- All administrators and regular users who have access to a specific Namespace in a cluster can also access:
    - Source code of all lambdas within this Namespace
    - Internal Docker registry that contains lambda images
