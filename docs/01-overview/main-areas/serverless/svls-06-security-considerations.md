---
title: Security considerations
---

To eliminate potential security risks when using Functions, bear in mind these few facts:

- Kyma provides base images for serverless runtimes. Those default runtimes are maintained with regards to commonly known security advisories. It is possible to use a custom runtime image (see this [tutorial](../../../03-tutorials/00-serverless/svls-13-override-runtime-image.md)). In such a case, you are responsible for security compliance and assessment of exploitability of any potential vulnerabilities of the custom runtime image.

- Kyma does not run any security scans against Functions and their images. Before you store any sensitive data in Functions, consider the potential risk of data leakage.

- Kyma does not define any authorization policies that would restrict Functions' access to other resources within the Namespace. If you deploy a Function in a given Namespace, it can freely access all events and APIs of services within this Namespace.

- All administrators and regular users who have access to a specific Namespace in a cluster can also access:

  - Source code of all Functions within this Namespace
  - Internal Docker registry that contains Function images
  - Secrets allowing the build Job to pull and push images from and to the Docker registry (in non-system Namespaces)
