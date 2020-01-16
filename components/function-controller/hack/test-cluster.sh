#!/usr/bin/env bash

# easy script to setup
# TODO add basic trap on err

set -eo pipefail

readonly KIND_VERSION="v0.5.1"
readonly STABLE_KUBERNETES_VERSION="v1.15.3"
readonly TEKTON_VERION="v0.7.0"
readonly KNATIVE_SERVING_VERSION="v0.8.0"

readonly TMP_DIR="$(mktemp -d)"
readonly TMP_BIN_DIR="${TMP_DIR}/bin"
mkdir -p "${TMP_BIN_DIR}"
export PATH="${TMP_BIN_DIR}:${PATH}"

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

kind::download_kind() {
  local -r kind_version="${1}"
  local -r host_os="${2}"
  local -r destination_dir="${3}"

  echo "Downloading kind in version ${kind_version}..."
  curl -LO "https://github.com/kubernetes-sigs/kind/releases/download/${kind_version}/kind-${host_os}-amd64" --fail \
      && chmod +x "kind-${host_os}-amd64" \
      && mv "kind-${host_os}-amd64" "${destination_dir}/kind"

  echo "Kind downloaded."
}

function kind::create_cluster {
    echo "Creating kind cluster"
    local -r image="kindest/node:${2}"

    kind create cluster --name "${1}" --image "${image}" --config "${SCRIPT_DIR}/cluster-config.yaml" --wait 3m

    local -r kubeconfig="$(kind get kubeconfig-path --name="${1}")"

    echo "export KUBECONFIG="$(kind get kubeconfig-path --name="fun-controller")""

    cp "${kubeconfig}" "${HOME}/.kube/config"
    kubectl cluster-info
    echo "Cluster created"
}

istio::download_istioctl(){
    local -r destination_dir="${1}"
    echo "Downloading istio"
    curl -L https://istio.io/downloadIstio | sh - \
    && chmod +x "istio-1.4.3/bin/istioctl" \
    && mv "istio-1.4.3/bin/istioctl" "${destination_dir}/istioctl" \
    && rm -rf "istio-1.4.3"
    echo "Downloaded istioctl"
}

istio::install(){
    istioctl verify-install
    istioctl manifest apply --skip-confirmation
}

tekton::install(){
    kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin
    kubectl apply -f "https://storage.googleapis.com/tekton-releases/pipeline/previous/${TEKTON_VERION}/release.yaml" --wait=true
}

cert-manager::install(){
    kubectl create namespace cert-manager
    kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.12.0/cert-manager.yaml
}

knative::install_serving(){
    # there's no guarantee that serving installs like this if the version is other than v0.8.0, so if
    # you change KNATIVE_SERVING_VERSION variable make sure the installation procedure didn't change
    kubectl apply --selector knative.dev/crd-install=true \
    --filename "https://github.com/knative/serving/releases/download/${KNATIVE_SERVING_VERSION}/serving.yaml" \
    --filename "https://github.com/knative/eventing/releases/download/${KNATIVE_SERVING_VERSION}/release.yaml" \
    --filename "https://github.com/knative/serving/releases/download/${KNATIVE_SERVING_VERSION}/monitoring.yaml" --wait=true || true

    kubectl apply --filename "https://github.com/knative/serving/releases/download/${KNATIVE_SERVING_VERSION}/serving.yaml" \
    --filename "https://github.com/knative/eventing/releases/download/${KNATIVE_SERVING_VERSION}/release.yaml" \
    --filename "https://github.com/knative/serving/releases/download/${KNATIVE_SERVING_VERSION}/monitoring.yaml" \
    --wait=true
}

main(){
    local -r kindClusterName="fun-controller"
    local -r imageName="function-controller"

    kind::download_kind "${KIND_VERSION}" "darwin" "${TMP_BIN_DIR}"
    istio::download_istioctl "${TMP_BIN_DIR}"
    kind::create_cluster "${kindClusterName}" "${STABLE_KUBERNETES_VERSION}"

    istio::install
    tekton::install

    knative::install_serving

    cert-manager::install

    kubectl create ns serverless-system
    docker build "${SCRIPT_DIR}/.." -t function-controller
    kind --name "${kindClusterName}" load docker-image "${imageName}:latest"

    # next -> wait for all pods to be ready ( watch kubectl get pods --all-namespaces)
    # make deploy

    # patch imagePullPolicy from Always to IfNotPresent to use local image
    #    kubectl -n serverless-system patch deployment function-controller-manager -p '
    #    apiVersion: extensions/v1beta1
    #    kind: Deployment
    #    metadata:
    #      name: function-controller-manager
    #      namespace: serverless-system
    #    spec:
    #        spec:
    #          containers:
    #          - args:
    #            image: function-controller:latest
    #            imagePullPolicy: IfNotPresent'
}

main