---
title: Security
type: Details
---

To eliminate potential security risks that the use of lambdas can involve, bear in mind these few facts: 

- Kyma doesn't run any security scans against functions and their images. Before you store any sensitive data in functions, consider the potential risk of data leakage.

- JSON Web Tokens (JWTs) used in Kyma are cluster-wide as they don't have the **scope** parameter defined. This means that if you expose your function and secure it with a JWT issued by Dex, you can use the token to validate access to all functions within the cluster.

- Kyma does not define any authorization policies that would restrict lambdas' access to other resources within the Namespace. If you deploy a lambda in a given Namespace, it can freely access all events and APIs of services within this Namespace.

- All administrators and regular users who have access to specific Namespaces in a cluster can access:
    - Source code of all lambdas within these Namespaces
    - Internal Docker registry that contains lambda images
