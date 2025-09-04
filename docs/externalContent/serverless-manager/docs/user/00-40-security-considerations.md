<!-- This document is a part of the Secure Development in the Kyma Environment section on HP -->
# Function Security

To eliminate potential security risks when using Functions, bear in mind these facts:

- By default, JSON Web Tokens (JWTs) issued by an OpenID Connect-compliant identity provider do not provide the scope parameter for Functions. This means that if you expose your Function and secure it with a JWT, you can use the token to validate access to all Functions within the cluster as well as other JWT-protected services.

- Kyma provides base images for serverless runtimes. Those default runtimes are maintained with regards to commonly known security advisories. It is possible to use a custom runtime image. For more information, see [Override Runtime Image](tutorials/01-110-override-runtime-image.md). In such a case, you are responsible for security compliance and assessment of exploitability of any potential vulnerabilities of the custom runtime image.

- Kyma does not run any security scans against Functions and their images. Before you store any sensitive data in Functions, consider the potential risk of data leakage.

- Kyma does not define any authorization policies that would restrict Functions' access to other resources within the namespace. If you deploy a Function in a given namespace, it can freely access all events and APIs of services within this namespace.

- Since Kubernetes is [moving from PodSecurityPolicies to PodSecurity Admission Controller](https://kubernetes.io/docs/tasks/configure-pod-container/migrate-from-psp/), Kyma Functions require running in namespaces with the `baseline` Pod security level. The `restricted` level is not currently supported due to the requirements of the Function building process.

- The Kyma Serverless components can run with the PodSecurity Admission Controller support in the `restricted` Pod security level when using an external registry. When the Internal Docker Registry is enabled, the Internal Registry DaemonSet requires elevated privileges to function correctly, exceeding the limitations of both the `restricted` and `baseline` levels.

- All administrators and regular users who have access to a specific namespace in a cluster can also access:

  - Source code of all Functions within this namespace
  - Internal Docker registry that contains Function images
  - Secrets allowing the build Job to pull and push images from and to the Docker registry (in non-system namespaces)
  