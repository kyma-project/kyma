#!/usr/bin/env bash

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

function installHelmOperator {
  helm repo add fluxcd https://charts.fluxcd.io
  helm install  helm-operator fluxcd/helm-operator \
    --namespace flux \
    --set helm.versions=v3 \
	  --set --git-timeout=10m \
	  --create-namespace

	bash ${CURRENT_DIR}/is-ready.sh flux app helm-operator
}

#install cluster-essentials, testing and istio
function installPrerequisites {

  kubectl apply -f ${CURRENT_DIR}/../resources/helm-releases/prerequisite-releases/cluster-essentials.yaml
  waitForRelease cluster-essentials kyma-system

  kubectl apply -f ${CURRENT_DIR}/../resources/helm-releases/prerequisite-releases/testing.yaml
  waitForRelease testing kyma-system

  kubectl apply -f ${CURRENT_DIR}/../resources/helm-releases/prerequisite-releases/istio.yaml
  waitForRelease istio istio-system

  kubectl apply -f ${CURRENT_DIR}/../resources/helm-releases/prerequisite-releases/xip-patch.yaml
  waitForRelease xip-patch kyma-installer
}

function installKymaComponents {
  kubectl apply -f ${CURRENT_DIR}/../resources/helm-releases/
}

function createNamespaces {
  for ns in  "kyma-installer" "kyma-system" "istio-system" "knative-serving" "knative-eventing" "kyma-integration"
  do
    kubectl create ns $ns;
  done
}

function waitForRelease {
  while :
  do
    if [ "$(kubectl get helmreleases.helm.fluxcd.io "$1" -n "$2" -o jsonpath='{.status.releaseStatus}')" = "deployed" ]
    then
      echo "$1 is deployed :)"
      break
    else
      echo "$1 is not yet deployed -  waiting 5s..."
      sleep 5
    fi
  done
}

if [[ ! $(kubectl get deployments.apps -n flux helm-operator) ]]; then
  installHelmOperator
fi

createNamespaces
kubectl label ns kyma-installer istio-injection=disabled --overwrite
installPrerequisites
installKymaComponents