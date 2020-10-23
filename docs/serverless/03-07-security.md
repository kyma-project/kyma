---
title: Security
type: Details
---

To eliminate potential security risks when using Functions, bear in mind these few facts:

- Kyma does not run any security scans against Functions and their images. Before you store any sensitive data in Functions, consider the potential risk of data leakage.

- By default, JSON Web Tokens (JWTs) issued by Dex do not provide the **scope** parameter for Functions. This means that if you expose your Function and secure it with a JWT, you can use the token to validate access to all Functions within the cluster.

- Kyma does not define any authorization policies that would restrict Functions' access to other resources within the Namespace. If you deploy a Function in a given Namespace, it can freely access all events and APIs of services within this Namespace.

- All administrators and regular users who have access to a specific Namespace in a cluster can also access:

    - Source code of all Functions within this Namespace
    - Internal Docker registry that contains Function images
    - Secrets allowing the build Job to pull and push images from and to the Docker registry (in non-system Namespaces)
