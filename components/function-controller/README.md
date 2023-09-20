# Function Controller

The Function Controller is a Kubernetes controller that enables Kyma to manage Function resources. It uses Kubernetes Jobs, Deployments, Services, and HorizontalPodAutoscalers (HPA) under the hood.

## Prerequisites

The Function Controller requires the following components to be installed:

- [Istio](https://github.com/istio/istio/releases) (v1.4.3)
- [Docker registry](https://github.com/docker/distribution) (v2.7.1)

## Development

To develop the Function Controller, you need:
- [libgit2-dev](https://libgit2.org/) (v1.5)
- [controller-gen](https://github.com/kubernetes-sigs/controller-tools/releases/tag/v0.6.2) (v0.6.2)
- <!-- markdown-link-check-disable-line -->[kustomize](https://github.com/kubernetes-sigs/kustomize/releases/tag/kustomize%2Fv4.5.7) (v4.5.7)

To develop the component, use the formulae declared in the [Makefile](./Makefile).

To run the manager and set up envs correctly, source [controller.env](./hack/controller.env). 
You can customize the configuration by editing files in [hack](./hack) dir. 


### Environment variables

#### The Function Controller uses these environment variables:

| Variable                                                  | Description                                                                                                                                                                                                                                                                                                  | Default value                                                                                                                                          |
| --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **APP_METRICS_ADDRESS**                                   | Address on which controller metrics are exposed                                                                                                                                                                                                                                                              | `:8080`                                                                                                                                                |
| **APP_LEADER_ELECTION_ENABLED**                           | Field that enables one instance of the Function Controller to manage the traffic among all instances                                                                                                                                                                                                         | `false`                                                                                                                                                |
| **APP_LEADER_ELECTION_ID**                                | Name of the ConfigMap that specifies the main instance of the Function Controller that manages the traffic among all instances                                                                                                                                                                               | `serverless-controller-leader-election-helper`                                                                                                         |
| **APP_KUBERNETES_BASE_NAMESPACE**                         | Name of the Namespace with the serverless configuration (such as runtime, Secret and service account for the Docker registry) propagated to other Namespaces                                                                                                                                                 | `kyma-system`                                                                                                                                          |
| **APP_KUBERNETES_EXCLUDED_NAMESPACES**                    | List of Namespaces to which serverless configuration is not propagated                                                                                                                                                                                                                                       | `istio-system,knative-eventing,kube-node-lease,kube-public,kube-system,kyma-installer,kyma-system,natss`                |
| **APP_KUBERNETES_CONFIG_MAP_REQUEUE_DURATION**            | Period of time after which the ConfigMap Controller refreshes the status of a ConfigMap                                                                                                                                                                                                                      | `1m`                                                                                                                                                   |
| **APP_KUBERNETES_SECRET_REQUEUE_DURATION**                | Period of time after which the Secret Controller refreshes the status of a Secret                                                                                                                                                                                                                            | `1m`                                                                                                                                                   |
| **APP_KUBERNETES_SERVICE_ACCOUNT_REQUEUE_DURATION**       | Period of time after which the ServiceAccount Controller refreshes the status of a ServiceAccount                                                                                                                                                                                                            | `1m`                                                                                                                                                   |
| **APP_FUNCTION_IMAGE_REGISTRY_DOCKER_CONFIG_SECRET_NAME** | Name of the secret that contains hashed credentials to the Docker registry                                                                                                                                                                                                                                   | `serverless-image-pull-secret`                                                                                                                         |
| **APP_FUNCTION_IMAGE_PULL_ACCOUNT_NAME**                  | Name of the service account that contains credentials to the Docker registry                                                                                                                                                                                                                                 | `serverless`                                                                                                                                           |
| **APP_FUNCTION_REQUEUE_DURATION**                         | Period of time after which the Function Controller refreshes the status of a Function CR                                                                                                                                                                                                                     | `1m`                                                                                                                                                   |
| **APP_FUNCTION_BUILD_REQUESTS_CPU**                       | Minimum amount of CPU assigned to the Job to build a Function image. See [this](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu) document for available values.                                                                                         | `350m`                                                                                                                                                 |
| **APP_FUNCTION_BUILD_REQUESTS_MEMORY**                    | Minimum amount of memory assigned to the Job to build a Function image. See [this](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu) document for available values.                                                                                      | `750mi`                                                                                                                                                |
| **APP_FUNCTION_BUILD_LIMITS_CPU**                         | Maximum amount of CPU assigned to the Job to build a Function image. See [this](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu) document for available values.                                                                                         | `1`                                                                                                                                                    |
| **APP_FUNCTION_BUILD_LIMITS_MEMORY**                      | Maximum amount of memory assigned to the Job to build a Function image. See [this](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/#meaning-of-cpu) document for available values.                                                                                      | `1Gi`                                                                                                                                                  |
| **APP_FUNCTION_BUILD_RUNTIME_CONFIG_MAP_NAME**            | Name of the ConfigMap that contains the runtime definition                                                                                                                                                                                                                                                   | `dockerfile-nodejs18`                                                                                                                                  |
| **APP_FUNCTION_BUILD_EXECUTOR_ARGS**                      | List of arguments passed to the Kaniko executor                                                                                                                                                                                                                                                              | `--insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true`                                                                     |
| **APP_FUNCTION_BUILD_EXECUTOR_IMAGE**                     | Full name of the Kaniko executor image used for building Function images and pushing them to the Docker registry                                                                                                                                                                                             | `gcr.io/kaniko-project/executor:v0.22.0`                                                                                                               |
| **APP_FUNCTION_BUILD_REPOFETCHER_IMAGE**                  | Full name of the Repo-Fetcher init container used for cloning repository for the Kaniko executor                                                                                                                                                                                                             | `europe-docker.pkg.dev/kyma-project/prod/function-build-init:305bee60`                                                                                                  |
| **APP_FUNCTION_BUILD_MAX_SIMULTANEOUS_JOBS**              | Maximum number of build jobs running simultaneously                                                                                                                                                                                                                                                            | `5`                                                                                                                                                  |
| **APP_FUNCTION_DOCKER_INTERNAL_SERVER_ADDRESS**           | Internal server address of the Docker registry                                                                                                                                                                                                                                                               | `serverless-docker-registry.kyma-system.svc.cluster.local:5000`                                                                                        |
| **APP_FUNCTION_DOCKER_REGISTRY_ADDRESS**                  | External address of the Docker registry                                                                                                                                                                                                                                                                      | `registry.kyma.local`                                                                                                                                  |
| **APP_FUNCTION_TARGET_CPU_UTILIZATION_PERCENTAGE**        | Average CPU usage of all the Pods in a given Deployment. It is represented as a percentage of the overall requested CPU. If the CPU consumption is higher or lower than this limit, HorizontalPodAutoscaler (HPA) scales the Deployment and increases or decreases the number of Pod replicas accordingly. | `50`                                                                                                                                                     |

#### The Webhook uses these environment variables:

| Variable                                  | Description                                                                               | Default value        |
| ----------------------------------------- | ----------------------------------------------------------------------------------------- | -------------------- |
| **SYSTEM_NAMESPACE**                      | Namespace which contains the ServiceAccount and the Secret                               | `kyma-system`        |
| **WEBHOOK_SERVICE_NAME**                  | Name of the ServiceAccount which is used by the webhook server                           | `serverless-webhook` |
| **WEBHOOK_SECRET_NAME**                   | Name of the Secret which contains the certificate is used to register the webhook server  | `serverless-webhook` |
| **WEBHOOK_PORT**                          | Port on which the webhook server are exposed                                              | `8443`               |
| **WEBHOOK_VALIDATION_MIN_REQUEST_CPU**    | Minimum amount of requested the limits and requests CPU to pass through the validation    | `10m`                |
| **WEBHOOK_VALIDATION_MIN_REQUEST_MEMORY** | Minimum amount of requested the limits and requests memory to pass through the validation | `16Mi`               |
| **WEBHOOK_VALIDATION_MIN_REPLICAS_VALUE** | Minimum amount of replicas to pass through the validation                                 | `1`                  |
| **WEBHOOK_VALIDATION_RESERVED_ENVS**      | List of reserved envs                                                                     | `{}`                 |
| **WEBHOOK_DEFAULTING_REQUEST_CPU**        | Value of the request CPU which webhook should set if origin equals null                   | `50m`                |
| **WEBHOOK_DEFAULTING_REQUEST_MEMORY**     | Value of the request memory which webhook should set if origin equals null                | `64Mi`               |
| **WEBHOOK_DEFAULTING_LIMITS_CPU**         | Value of the limits CPU which webhook should set if origin equals null                    | `100m`               |
| **WEBHOOK_DEFAULTING_LIMITS_MEMORY**      | Value of the limits memory which webhook should set if origin equals null                 | `128Mi`              |
| **WEBHOOK_DEFAULTING_MINREPLICAS**        | Value of the minReplicas which webhook should set if origin equals null                   | `1`                  |
| **WEBHOOK_DEFAULTING_MAXREPLICAS**        | Value of the maxReplicas which webhook should set if origin equals null                   | `1`                  |

## Troubleshooting


### Symptom

Function Controller tests keep failing with such an error message:
`error: Invalid libgit2 version; this git2go supports libgit2 between vA.B.C and vX.Y.Z`

### Cause

Function Controller tests are failing due to the wrong version of the libgit2 binary. The required version of the binary is 1.1.

### Remedy

Build and install the libgit2 binary required by the Function Controller on macOS. Follow these steps:

1. Navigate to the Function Controller's root directory and verify the version of git2go:

   ```bash
   $ cat go.mod | grep git2go
   ```
   You should get a result similar to this example:

   ```bash
   github.com/libgit2/git2go/v31 v31.4.14
   ```
2. Go to the [git2go page](https://github.com/libgit2/git2go#git2go) to check which version of libgit2 you must use. For example, for git2go v34, use libigit2 in version 1.5.
   
3. Clone the `libgit2` repository:

   ```bash
   git clone https://github.com/libgit2/libgit2.git
   ```
4. Check out the sources. In this example, the sources are for git2go v31:
   ```bash
   git checkout v1.1.0
   ```
5. Build and install the libgit2 binary:
   ```bash
   cmake -DCMAKE_OSX_ARCHITECTURES="x86_64" .
   make install
   ```
   