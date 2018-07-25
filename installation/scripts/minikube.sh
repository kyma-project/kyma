#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

MINIKUBE_DOMAIN=""
MINIKUBE_VERSION=0.28.2
KUBERNETES_VERSION=1.10.0
VM_DRIVER="hyperkit"

source $CURRENT_DIR/utils.sh

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --vm-driver)
            VM_DRIVER="$2"
            shift # past argument
            shift # past value
            ;;
        --domain)
            MINIKUBE_DOMAIN="$2"
            shift # past argument
            shift # past value
            ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

echo "
################################################################################
# Minikube setup with driver ${VM_DRIVER}
################################################################################
"

function initializeMinikubeConfig() {
    # Disable default nginx ingress controller
    minikube config unset ingress
}

function uploadDexTlsCertForApiserver() {
	local TLS_FILE="$CURRENT_DIR/../resources/local-tls-certs.yaml"

	local TLS_CERT=$(cat $TLS_FILE | grep 'tls.crt' | sed 's/^.*: //')

	log "Parsing DEX TLS certificate secret..." green

    local crt=$(echo $TLS_CERT | base64 --decode)

    # cert saved for localkube... for kubeadm (--bootstrapper=kubeadm) location will be different
    local crtDir="$HOME/.minikube/files"

    if [ ! -d "$crtDir" ]; then
        mkdir -p "${crtDir}"
    fi

    local crtFile="$crtDir/dex-ca.crt"

    log "Saving DEX TLS certificate in the container file: ${crtFile}..." green

    echo -e "${crt}" > "${crtFile}"

    log "DEX TLS certificate saved in the container: ${crtFile}" green
}

#TODO refactor to use minikube status!
function waitForMinikubeToBeUp() {
    set +o errexit

    log "Waiting for minikube to be up..." green

	LIMIT=15
    COUNTER=0

    while [ ${COUNTER} -lt ${LIMIT} ] && [ -z "$STATUS" ]; do
      (( COUNTER++ ))
      log "Keep calm, there are $LIMIT possibilities and so far it is attempt number $COUNTER" green
      STATUS="$(kubectl get namespaces || :)"
      sleep 1
    done

    # In case apiserver is not available get localkube logs
    if [[ -z "$STATUS" ]] && [[ "$VM_DRIVER" = "none" ]]; then
      cat /var/lib/localkube/localkube.err
    fi

    set -o errexit

    log "Minikube is up" green
}

# After upgrade to minikube v0.24.1, our CI plans (use dind to start
# kyma) are not able to pull images from internet
function fixDindMinikubeIssue() {
    echo "nameserver 8.8.8.8" >> /etc/resolv.conf
}

function checkIfMinikubeIsInitialized() {
    local status=$(minikube status --format "{{.MinikubeStatus}}")
    if [ -n "${status}" ]; then
        log "Minikube is already initialized" red
        read -p "Do you want to remove previous minikube cluster [y/N]: " deleteMinikube
        if [ "${deleteMinikube}" == "y" ]; then
            minikube delete
        else
            log "Starting minikube cancelled" red
            exit -1
        fi
    fi
}

function checkMinikubeVersion() {
    local version=$(minikube version | awk '{print  $3}')

    if [[ "${version}" != *"${MINIKUBE_VERSION}"* ]]; then
        echo "Your minikube is in v${version}. v${MINIKUBE_VERSION} is supported version of minikube. Install supported version!"
        exit -1
    fi
}

function addDevDomainsToEtcHosts() {
    local hostnames=$1
    local minikubeIP=$(minikube ip)

    log "Minikube IP address: ${minikubeIP}" green

    if [[ "$VM_DRIVER" != "none" ]]; then
        log "Adding ${hostnames} to /etc/hosts on Minikube" yellow
        minikube ssh "echo \"127.0.0.1 ${hostnames}\" | sudo tee -a /etc/hosts"

        # Delete old host alias
        case `uname -s` in
            Darwin)
                sudo sed -i '' "/${MINIKUBE_DOMAIN}/d" /etc/hosts
                ;;
            *)
                sudo sed -i  "/${MINIKUBE_DOMAIN}/d" /etc/hosts
                ;;
        esac
    fi

    log "Adding ${hostnames} to /etc/hosts on localhost" yellow
	local hostAlias="${minikubeIP} ${hostnames}"

    #Set new host alias
    echo ${hostAlias} | sudo tee -a /etc/hosts > /dev/null

	log "Domain added to /etc/hosts" green
}

function start() {
    checkMinikubeVersion

    checkIfMinikubeIsInitialized

    initializeMinikubeConfig

    uploadDexTlsCertForApiserver

    if [[ "$VM_DRIVER" = "none" ]]; then
        fixDindMinikubeIssue
    fi

    minikube start \
    --memory 8192 \
    --cpus 4 \
    --extra-config=apiserver.Authorization.Mode=RBAC \
    --extra-config=apiserver.Authentication.OIDC.IssuerURL="https://dex.${MINIKUBE_DOMAIN}" \
    --extra-config=apiserver.Authentication.OIDC.CAFile=/dex-ca.crt \
    --extra-config=apiserver.Authentication.OIDC.ClientID=kyma-client \
    --extra-config=apiserver.Authentication.OIDC.UsernameClaim=email \
    --extra-config=apiserver.Authentication.OIDC.GroupsClaim=groups \
    --extra-config=apiserver.GenericServerRunOptions.CorsAllowedOriginList=".*" \
    --extra-config=controller-manager.ClusterSigningCertFile="/var/lib/localkube/certs/ca.crt" \
	--extra-config=controller-manager.ClusterSigningKeyFile="/var/lib/localkube/certs/ca.key" \
    --extra-config=apiserver.Admission.PluginNames="Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota,PodPreset,PersistentVolumeLabel" \
    --kubernetes-version=v$KUBERNETES_VERSION \
    --vm-driver=$VM_DRIVER \
    --feature-gates="MountPropagation=false" \
    -b=localkube

    waitForMinikubeToBeUp

    # Adding domains to /etc/hosts files
    addDevDomainsToEtcHosts "apiserver.${MINIKUBE_DOMAIN} console.${MINIKUBE_DOMAIN} catalog.${MINIKUBE_DOMAIN} instances.${MINIKUBE_DOMAIN} dex.${MINIKUBE_DOMAIN} docs.${MINIKUBE_DOMAIN} lambdas-ui.${MINIKUBE_DOMAIN} ui-api.${MINIKUBE_DOMAIN} minio.${MINIKUBE_DOMAIN} jaeger.${MINIKUBE_DOMAIN} grafana.${MINIKUBE_DOMAIN}  configurations-generator.${MINIKUBE_DOMAIN} gateway.${MINIKUBE_DOMAIN} connector-service.${MINIKUBE_DOMAIN}"
}

start

