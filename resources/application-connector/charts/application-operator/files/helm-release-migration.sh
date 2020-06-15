#!/usr/bin/env bash

set -e
HELM_2_BINARY=$(which helm)
HELM_3_BINARY=$(which helm3)
SECRET_NAME="helm-secret"
OVERRIDES_NAMESPACE="kyma-installer"

APPLICATION_RELEASES_NAMESPACE="kyma-integration"

echo "---> Install requirements"

apk add git jq

echo "---> Get HELM_2 certs"
${HELM_2_BINARY} init -c
kubectl get -n "${OVERRIDES_NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\\.helm\\.ca\\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
kubectl get -n "${OVERRIDES_NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\\.helm\\.tls\\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
kubectl get -n "${OVERRIDES_NAMESPACE}" secret "${SECRET_NAME}" -o jsonpath="{.data['global\\.helm\\.tls\\.key']}" | base64 --decode > "$(helm home)/key.pem"

echo "---> Setup Helm 3"
if [[ ! $(${HELM_3_BINARY} plugin list | grep '2to3') ]]; then
    echo "---> Get migration plugin"
    ${HELM_3_BINARY} plugin install https://github.com/helm/helm-2to3.git

    echo "---> Migrate config files"
    yes | ${HELM_3_BINARY} 2to3 move config
fi

echo "---> Get current releases "

${HELM_2_BINARY} ls --tls --all --output json | jq '.Releases[] | .Name + " " + .Namespace + " " + .Chart' | tr -d '"' > helm2-all-releases

echo "---> Migrate Gateway and Application releases"
while read line; do
    release=$(echo $line | cut -d " " -f1)
    ns=$(echo $line | cut -d " " -f2)
    chart=$(echo $line | cut -d " " -f3)

    if [[ "$chart" = "gateway-0.0.1" ]]; then
        echo "------> Migrating Gateway release $release"
        if [[ $(${HELM_3_BINARY} get all ${release} -n ${ns}) ]]; then
            echo "------> Release ${release} in ns ${ns} already migrated!"
            ${HELM_2_BINARY} delete --purge ${release} --tls
        else
            ${HELM_3_BINARY} 2to3 convert ${release}
            ${HELM_2_BINARY} delete --purge ${release} --tls
        fi
    fi

    if [[ "$chart" = "application-0.0.1" ]]; then
        echo "------> Migrating Application release $release"
        if [[ $(${HELM_3_BINARY} get all ${release} -n ${ns}) ]]; then
            echo "------> Release ${release} in ns ${ns} already migrated!"
        else
            ${HELM_3_BINARY} 2to3 convert ${release}
        fi
    fi

done < helm2-all-releases
