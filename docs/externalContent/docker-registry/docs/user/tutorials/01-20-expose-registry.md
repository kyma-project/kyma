# Expose Docker Registry

This tutorial shows how you can expose the registry to the outside of the cluster with Istio.

## Prerequsities

* [kubectl](https://kubernetes.io/docs/tasks/tools/)
* [Docker](https://www.docker.com/)

## Steps

1. Expose the registry service by changing the **spec.externalAccess.enabled** flag to `true`:

    ```bash
    kubectl apply -n kyma-system -f - <<EOF
    apiVersion: operator.kyma-project.io/v1alpha1
    kind: DockerRegistry
    metadata:
      name: default
      namespace: kyma-system
    spec:
      externalAccess:
        enabled: true
    EOF
    ```

   Once the DockerRegistry CR becomes `Ready`, you see a Secret name used as `ImagePullSecret` when scheduling workloads in the cluster.

    ```yaml
    ...
    status:
      externalAccess:
        enabled: "True"
        ...
        secretName: dockerregistry-config-external
    ```

2. Log in to the registry using the docker-cli:

    ```bash
    export REGISTRY_USERNAME=$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.username} | base64 -d)
    export REGISTRY_PASSWORD=$(kubectl get secrets -n kyma-system dockerregistry-config-external -o jsonpath={.data.password} | base64 -d)
    export REGISTRY_ADDRESS=$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.externalAccess.pushAddress})
    docker login -u "${REGISTRY_USERNAME}" -p "${REGISTRY_PASSWORD}" "${REGISTRY_ADDRESS}"
    ```

3. Rename the image to contain the registry address:

    ```bash
    export IMAGE_NAME=<IMAGE_NAME> # put your image name here
    docker tag "${IMAGE_NAME}" "${REGISTRY_ADDRESS}/${IMAGE_NAME}"
    ```

4. Push the image to the registry:

    ```bash
    docker push "${REGISTRY_ADDRESS}/${IMAGE_NAME}"
    ```

5. Create a Pod using the image from Docker Registry:

    ```bash

    export REGISTRY_INTERNAL_PULL_ADDRESS=$(kubectl get dockerregistries.operator.kyma-project.io -n kyma-system default -ojsonpath={.status.internalAccess.pullAddress})
    kubectl run my-pod --image="${REGISTRY_INTERNAL_PULL_ADDRESS}/${IMAGE_NAME}" --overrides='{ "spec": { "imagePullSecrets": [ { "name": "dockerregistry-config" } ] } }'
    ```
