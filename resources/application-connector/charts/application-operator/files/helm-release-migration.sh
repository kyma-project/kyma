#!/usr/bin/env bash

set -e

exit 0

HELM_2_BINARY=$(which helm)
HELM_3_BINARY=$(which helm3)
SECRET_NAME="helm-secret"
NAMESPACE="kyma-installer"

echo "---> Install requirements"

apk add git jq

echo "---> Get HELM_2 certs"
${HELM_2_BINARY} init -c --skip-refresh

if [[ $(kubectl get -n kube-system deploy tiller-deploy -o name) ]]; then
    if [[ $(kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" -o name) ]]; then
        kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\\.helm\\.ca\\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
        kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\\.helm\\.tls\\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
        kubectl get -n "${NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\\.helm\\.tls\\.key']}" | base64 --decode > "$(helm home)/key.pem"
    else
        exit 0
    fi
else
    echo "------> No Tiller deployment found, exiting gracefully"
    exit 0
fi

echo "---> Setup Helm 3"
if [[ ! $(${HELM_3_BINARY} plugin list | grep '2to3') ]]; then
    echo "---> Get migration plugin"
    ${HELM_3_BINARY} plugin install https://github.com/helm/helm-2to3.git

    echo "---> Migrate config files"
    yes | ${HELM_3_BINARY} 2to3 move config
fi

echo "---> Get current releases "
set -o pipefail
${HELM_2_BINARY} ls --tls --all --output json | jq '.Releases[] | .Name + " " + .Namespace + " " + .Chart' | tr -d '"' > helm2-all-releases
set +o pipefail

echo "---> Migrate Gateway and Application releases"
while read line; do
    release=$(echo $line | cut -d " " -f1)
    ns=$(echo $line | cut -d " " -f2)
    chart=$(echo $line | cut -d " " -f3)

    case "$chart" in
        "gateway-0.0.1") rtype="Gateway" ;;
        "application-0.0.1") rtype="Application" ;; 
        *) continue ;;
    esac

    echo "------> Migrating ${rtype} release $release"
    if [[ $(${HELM_3_BINARY} get all ${release} -n ${ns}) ]]; then
        echo "------> Release ${release} in ns ${ns} already migrated!"
        yes | ${HELM_3_BINARY} 2to3 cleanup --name ${release}
    else
        ${HELM_3_BINARY} 2to3 convert ${release}
        yes | ${HELM_3_BINARY} 2to3 cleanup --name ${release}
    fi 

done < helm2-all-releases
