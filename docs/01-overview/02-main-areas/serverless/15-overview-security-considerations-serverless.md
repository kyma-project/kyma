---
title: Security considerations
---

To eliminate potential security risks when using Functions, bear in mind these few facts:

- Kyma does not run any security scans against Functions and their images. Before you store any sensitive data in Functions, consider the potential risk of data leakage.

- Kyma does not define any authorization policies that would restrict Functions' access to other resources within the Namespace. If you deploy a Function in a given Namespace, it can freely access all events and APIs of services within this Namespace.

- All administrators and regular users who have access to a specific Namespace in a cluster can also access:

    - Source code of all Functions within this Namespace
    - Internal Docker registry that contains Function images
    - Secrets allowing the build Job to pull and push images from and to the Docker registry (in non-system Namespaces)

Serverless Functions are adapted to run in a non-privileged mode. If you [enable Pod Security Policies](https://kubernetes.io/docs/concepts/policy/pod-security-policy/#enabling-pod-security-policies) in your cluster, Functions will have the following security measures set in place:

- The root filesystem is set to read-only mode, apart from the `/tmp` directory which is read-write.
- The Function's container is non-privileged and runs as a non-root user.
- All [Linux capabilities](https://kubernetes.io/docs/concepts/policy/pod-security-policy/#capabilities) are dropped.
