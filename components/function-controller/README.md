# Function Controller

The Knative Function Controller is a Kubernetes controller that enables Kyma to manage Function resources. It uses Knative Build and Knative Serving under the hood.

## Prerequisites

The Function Controller requires the following components to be installed:

- Knative Build (v0.6.0)
- Knative Serving (v0.6.1)
- Istio (v1.0.7)

## Installation

### Set up the environment

Follow these steps to prepare the environment you will use to deploy the Controller.

1. Export the following environment variables:

    | Variable        | Description | Sample value |
    | --------------- | ----------- |--------------|
    | **IMG** | The full image name the Function Controller will be tagged with. | `gcr.io/my-project/function-controller` for GCR, `my-user/function-controller` for Docker Hub |
    | **FN_REGISTRY** | The URL of the container registry Function images will be pushed to. Used for authentication. | `https://gcr.io/` for GCR, `https://index.docker.io/v1/` for Docker Hub |
    | **FN_REPOSITORY** | The name of the container repository Function images will be pushed to. | `gcr.io/my-project` for GCR, `my-user` for Docker Hub |
    | **FN_NAMESPACE** | The Namespace where Functions are deployed. | `sample-namespace` |

    See the example:

    ```bash
    export IMG=my-docker-user/function-controller
    export FN_REGISTRY=https://index.docker.io/v1/
    export FN_REPOSITORY=my-docker-user
    export FN_NAMESPACE=my-functions
    ```

2. Create the `serverless-system` Namespace you will deploy the controller to.

    ```bash
    kubectl create namespace serverless-system
    ```

3. Create the following configuration for the controller. It contains a list of supported runtimes as well as the container repository referenced by the **FN_REPOSITORY** environment variable, which you will create a Secret for in the next steps.

    ```bash
    cat <<EOF | kubectl -n serverless-system apply -f -
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: fn-config
    data:
      serviceAccountName: function-controller-build
      dockerRegistry: ${FN_REPOSITORY}
      runtimes: |
        - ID: nodejs8
          dockerFileName: dockerfile-nodejs-8
        - ID: nodejs6
          dockerFileName: dockerfile-nodejs-6
    EOF
    ```

4. Create the Namespace defined previously by the **FN_NAMESPACE** environment variable. Function will be deployed to it.

    ```bash
    kubectl create namespace ${FN_NAMESPACE}
    ```

5. Functions require Dockerfiles to be created for each supported runtime. The following manifest contains Dockerfiles for Node.js runtimes.

    ```bash
    kubectl apply -n ${FN_NAMESPACE} -f config/dockerfiles.yaml
    ```

6. Before you create Functions, it is necessary to create the `registry-credentials` Secret, which contains credentials to the Docker registry defined by the **FN_REGISTRY** environment variable. Knative Build uses this Secret to push the images it builds for the deployed Functions. The corresponding `function-controller-build` ServiceAccount was referenced inside the controller configuration in step 3.

    ```bash
    reg_username=<container registry username>
    reg_password=<container registry password>

    cat <<EOF | kubectl -n ${FN_NAMESPACE} apply -f -
    ---
    apiVersion: v1
    kind: Secret
    type: kubernetes.io/basic-auth
    metadata:
      name: registry-credentials
      annotations:
        build.knative.dev/docker-0: ${FN_REGISTRY}
    stringData:
      username: ${reg_username}
      password: ${reg_password}
    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: function-controller-build
    secrets:
    - name: registry-credentials
    EOF
    ```

### Deploy the controller

To deploy the Function Controller to the `serverless-system` Namespace run

    ```bash
    make deploy
    ```
   
   This runs `ko` to build your image and push it to your configured docker repository. 
  See [google/ko](https://github.com/google/ko) for more details about `ko`

## Usage

### Run tests

Use the following `make` target to test the deployed Function Controller.

```bash
make test
```

### Create a sample Hello World Function

Follow the steps below to create a sample Function.

1. Apply the following Function manifest:

    ```bash
    kubectl apply -f config/samples/serverless_v1alpha1_function.yaml
    ```

2. Ensure the Function was created:

    ```bash
    kubectl get functions
    ```

3. Check the status of the build:

    ```bash
    kubectl get builds.build.knative.dev
    ```

4. Check the status of the Knative Serving service:

    ```bash
    kubectl get services.serving.knative.dev
    ```

5. Access the Function:

    <div tabs name="installation">

      <details>
      <summary>Minikube</summary>

      ```bash
      FN_DOMAIN="$(kubectl get ksvc demo --output 'jsonpath={.status.domain}')"
      FN_PORT="$(kubectl get svc istio-ingressgateway -n istio-system --output 'jsonpath={.spec.ports[?(@.port==80)].nodePort}')"
      curl -v -H "Host: ${FN_DOMAIN}" http://$(minikube ip):${FN_PORT}
      ```
      </details>

      <details>
      <summary>Remote cluster</summary>

      ```bash
      FN_DOMAIN="$(kubectl get ksvc demo --output 'jsonpath={.status.domain}')"
      curl -kD- "https://${FN_DOMAIN}"
      ```
      </details>

    </div>
