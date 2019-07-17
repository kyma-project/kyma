# Knative Function Controller

## Overview

The Knative Function Controller is a Kubernetes Controller that enable Kyma to manage Function resources.

It defines and handles a function Custom Resource Definition with the help of Knative Build and Knative Serving. Basically it is the serverless implementation based on knative.

>**NOTE:** Currently Kyma uses [Kubeless](https://github.com/kubeless/kubeless) as a default Serverless implementation.

## Requirements & Scaffolding

The code got scaffolded with `kubebuilder==1.0.8`.
Make sure to use `kustomize==1.0.10`. Otherwise you will get security errors.

The code has been scaffoled using the following commands:

```bash
kubebuilder init --domain kyma-project.io
kubebuilder create api --group runtime --version v1alpha1 --kind Function
kubebuilder alpha webhook --group runtime --version v1alpha1 --kind Function --type mutating
```

## WebHook Readme

The following section shows some links on webhooks and how the webhook mutates and validates the example functions.

### Links

- <https://book-v1.book.kubebuilder.io/beyond_basics/sample_webhook.html>
- <https://github.com/morvencao/kube-mutating-webhook-tutorial/blob/master/medium-article.md>

### Alternative

[Custom Resource Validation Schema](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#publish-validation-schema-in-openapi-v2)

### Examples

#### Mutation

Deploy Function:

```bash
kubectl apply -f config/samples/runtime_v1alpha1_function.yaml
```

Verify default values got set:

```bash
$ kubectl get functions.runtime.kyma-project.io function-sample -oyaml
...
spec:
  function: |
    module.exports = {
        main: function(event, context) {
          return 'Hello World'
        }
      }
  functionContentType: plaintext
  runtime: nodejs8
  size: S
  timeout: 180
```

#### Validation

Deploy Function:

```bash
$ kubectl apply -f config/samples/runtime_v1alpha1_function_invalid.yaml
Error from server (InternalError): error when creating "config/samples/runtime_v1alpha1_function_invalid.yaml": Internal error occurred: admission webhook "mutating-create-function.kyma-project.io" denied the request: runtime should be one of 'nodejs6,nodejs8'
```

## Development

### Test

```bash
make test
```

### Setup knative

start a beefy minikube

```bash
minikube start \
  --memory=12288 \
  --cpus=4 \
  --kubernetes-version=v1.12.0 \
  --vm-driver=hyperkit \
  --disk-size=30g \
  --extra-config=apiserver.enable-admission-plugins="LimitRanger,NamespaceExists,NamespaceLifecycle,ResourceQuota,ServiceAccount,DefaultStorageClass,MutatingAdmissionWebhook"
```

install istio

```bash
kubectl apply \
  --filename https://raw.githubusercontent.com/knative/serving/v0.5.2/third_party/istio-1.0.7/istio-crds.yaml &&
curl -L https://raw.githubusercontent.com/knative/serving/v0.5.2/third_party/istio-1.0.7/istio.yaml \
  | sed 's/LoadBalancer/NodePort/' \
  | kubectl apply --filename -
```

install knative

```bash
kubectl apply --selector knative.dev/crd-install=true \
--filename https://github.com/knative/serving/releases/download/v0.6.1/serving.yaml \
--filename https://github.com/knative/build/releases/download/v0.6.0/build.yaml \
--filename https://github.com/knative/eventing/releases/download/v0.6.1/release.yaml \
--filename https://github.com/knative/eventing-sources/releases/download/v0.6.0/eventing-sources.yaml \
--filename https://github.com/knative/serving/releases/download/v0.6.1/monitoring.yaml \
--filename https://raw.githubusercontent.com/knative/serving/v0.6.1/third_party/config/build/clusterrole.yaml

```

install knative part2

```bash
kubectl apply --filename https://github.com/knative/serving/releases/download/v0.6.0/serving.yaml --selector networking.knative.dev/certificate-provider!=cert-manager \
   --filename https://github.com/knative/build/releases/download/v0.6.0/build.yaml \
   --filename https://github.com/knative/eventing/releases/download/v0.6.0/release.yaml \
   --filename https://github.com/knative/eventing-sources/releases/download/v0.6.0/eventing-sources.yaml \
   --filename https://github.com/knative/serving/releases/download/v0.6.0/monitoring.yaml \
   --filename https://raw.githubusercontent.com/knative/serving/v0.6.0/third_party/config/build/clusterrole.yaml
```

### Local Deployment

#### Manager running locally

modify config/config.yaml to include your docker.io credentials (base64 encoded) and update the docker registry value to your docker.io username

Apply the configuration

```bash
kubectl apply -f config/config.yaml
```

Install the CRD to a local Kubernetes cluster:

```bash
make install
```

Run the controller on your machine:

```bash
make run
```

#### Manager running inside k8s cluster

This workflow needs to be used until [Come up with webhook developer workflow to test it locally #400](https://github.com/kubernetes-sigs/kubebuilder/issues/400) is fixed.

```bash
eval $(minikube docker-env)
make docker-build
make install
make deploy
```

### Prod Deployment

Uncomment `manager_image_patch_dev` in `kustomization.yaml`
Then run the following commands:

```bash
make install
make docker-push IMG=<e.g. index.docker.io/nachtmaar/runtime-controller>
make deploy
```

### Run the examples

Create sample function

```bash
kubectl apply -f config/samples/runtime_v1alpha1_function.yaml -n {NAMESPACE}
```

search for function

```bash
kubectl get functions -n {NAMESPACE}
```

```bash
kubectl get function -n {NAMESPACE}
```

```bash
kubectl get fcn -n {NAMESPACE}
```

access the function

```bash
	curl -v -H "Host: $(kubectl get ksvc sample --output 'jsonpath={.status.domain}' -n {NAMESPACE}" http://$(minikube ip):$(kubectl get svc istio-ingressgateway --namespace istio-system --output 'jsonpath={.spec.ports[?(@.port==80)].nodePort}')
```