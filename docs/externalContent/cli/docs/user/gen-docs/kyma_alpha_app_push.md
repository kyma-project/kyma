# kyma alpha app push

Push the application to the Kubernetes cluster.

## Synopsis

Use this command to push the application to the Kubernetes cluster.

```bash
kyma alpha app push [flags]
```

## Flags

```text
      --code-path string                   Path to the application source code directory
      --container-port int                 Port on which the application is exposed
      --dockerfile string                  Path to the Dockerfile
      --dockerfile-build-arg stringArray   Variables used while building an application from Dockerfile as args
      --dockerfile-context string          Context path for building Dockerfile (defaults to the current working directory)
      --expose                             Creates an APIRule for the app
      --image string                       Name of the image to deploy
      --image-pull-secret string           Name of the Kubernetes Secret with credentials to pull the image
      --istio-inject                       Enables Istio for the app
      --mount-config stringArray           Mounts ConfigMap content to the /bindings/configmap-<CONFIGMAP_NAME> path (default "[]")
      --mount-secret stringArray           Mounts Secret content to the /bindings/secret-<SECRET_NAME> path (default "[]")
      --name string                        Name of the app
      --namespace string                   Namespace where the app is deployed (default "default")
  -h, --help                               Help for the command
      --kubeconfig string                  Path to the Kyma kubeconfig file
      --show-extensions-error              Prints a possible error when fetching extensions fails
      --skip-extensions                    Skip fetching extensions from the cluster
```

## See also

* [kyma alpha app](kyma_alpha_app.md) - Manages applications on the Kubernetes cluster
