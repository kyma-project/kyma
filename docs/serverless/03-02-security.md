---
title: Security
type: Details
---

To eliminate potential security risks when using functions, bear in mind these few facts:

- Kyma does not run any security scans against functions and their images. Before you store any sensitive data in functions, consider the potential risk of data leakage.

- By default, JSON Web Tokens (JWTs) issued by Dex do not provide the **scope** parameter for functions. This means that if you expose your function and secure it with a JWT, you can use the token to validate access to all functions within the cluster.

- Kyma does not define any authorization policies that would restrict functions' access to other resources within the Namespace. If you deploy a function in a given Namespace, it can freely access all events and APIs of services within this Namespace.

- All administrators and regular users who have access to a specific Namespace in a cluster can also access:
    - Source code of all functions within this Namespace
    - Internal Docker registry that contains function images
