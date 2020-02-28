# Function Controller

The Function Controller is a Kubernetes controller that enables Kyma to manage Function resources. It uses Tekton Pipelines and Knative Serving under the hood.

> **CAUTION:** Functions work only in the `serverless` Namespace that is created as part of the `function-controller` chart. To avoid authentication issues, add a trusted certificate to your cluster. Read [here](https://kyma-project.io/docs/#installation-install-kyma-with-your-own-domain-generate-the-tls-certificate) how to obtain one.

## Prerequisites

The Function Controller requires the following components to be installed:

- [Tekton Pipelines](https://github.com/tektoncd/pipeline/releases) (v0.10.1)
- [Knative Serving](https://github.com/knative/serving/releases) (v0.12.1)
- [Istio](https://github.com/istio/istio/releases) (v1.4.3)
- [Cert Manager](https://github.com/jetstack/cert-manager) (v0.13.0)
- [Docker Registry](https://github.com/docker/distribution) (v2.7.1)

## Development

To develop the component, use the formulae declared in the [generic](/common/makefiles/generic-make-go.mk) and [component-specific](./Makefile) Makefiles. To run tests without the Makefile logic, use the `go test ./...` command.

### Environment variables

The Function Controller uses these environment variables:

| Variable                                        | Description                                                                                                                                                                                                                                | Default value                                                       |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------- |
| **BUILD_TIMEOUT**                               | Timeout after which building Tekton TaskRuns fails. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". See [this](https://golang.org/pkg/time/#ParseDuration) link for reference.                                             | `30m`                                                               |
| **CONTROLLER_CONFIGMAP**                        | Name of the ConfigMap containing information about available runtimes, function types, and default values                                                                                                                                  | `fn-config`                                                         |
| **CONTROLLER_CONFIGMAP_NS**                     | Namespace in which the ConfigMap resides.                                                                                                                                                                                                  | `default`                                                           |
| **CONTROLLER_DOCKER_REGISTRY_FQDN**             | Fully qualified domain name of the Docker Registry Service                                                                                                                                                                                 | `function-controller-docker-registry.kyma-system.svc.cluster.local` |
| **CONTROLLER_DOCKER_REGISTRY_PORT**             | Port of the Docker Registry Service through which communication is conducted                                                                                                                                                               | `5000`                                                              |
| **CONTROLLER_DOCKER_REGISTRY_EXTERNAL_ADDRESS** | Domain address through which connection to the Docker Registry is facilitated                                                                                                                                                              | `https://registry.kyma.local`                                       |
| **CONTROLLER_IMAGE_PULL_SECRET_NAME**           | Name of the image Pull Secret which contains hashed credentials to the Docker Registry                                                                                                                                                     | `regcred`                                                           |
| **CONTROLLER_TEKTON_REQUESTS_CPU**              | Minimum amount of CPU assigned to the TaskRun to build the lambda image. See [this](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu) document for available values.                   | `350m`                                                              |
| **CONTROLLER_TEKTON_REQUESTS_MEMORY**           | Minimum amount of Memory assigned to the TaskRun to build the lambda image. See [this](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu) document for available values. | `600Mi`                                                               |
| **CONTROLLER_TEKTON_LIMITS_CPU**                | Minimum amount of CPU assigned to the TaskRun to build the lambda image. See [this](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu) document for available values.  | `400m`                                                                |
| **CONTROLLER_TEKTON_LIMITS_MEMORY**             | Minimum amount of Memory assigned to the TaskRun to build the lambda image. See [this](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu) document for available values. | `700Mi`                                                              |

> **NOTE:** Tekton handles resources in a specific way depending on limit ranges. You can find more about it [here](https://github.com/tektoncd/pipeline/blob/master/docs/taskruns.md#limitranges).
