#!/usr/bin/env bash

echo "The minikube.sh script is deprecated and will be removed. Use Kyma CLI instead."

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCES_DIR="${CURRENT_DIR}/../resources"

MINIKUBE_DOMAIN=""
# Supported Minikube Versions: MINIKUBE_VERSION_MIN (inclusive) up to MINIKUBE_VERSION_MAX (exclusive)
MINIKUBE_VERSION=1.3.1
KUBERNETES_VERSION=1.14.6
KUBECTL_CLI_VERSION=1.14.6
VM_DRIVER=hyperkit
DISK_SIZE=30g
MEMORY=8192

source $CURRENT_DIR/utils.sh

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --disk-size)
            checkInputParameterValue "$2"
            DISK_SIZE="$2"
            shift
            shift
            ;;
        --vm-driver)
            checkInputParameterValue "$2"
            VM_DRIVER="$2"
            shift # past argument
            shift # past value
            ;;
        --memory)
            checkInputParameterValue "$2"
            MEMORY="$2"
            shift
            shift
            ;;
        --domain)
            checkInputParameterValue "$2"
            MINIKUBE_DOMAIN="$2"
            shift # past argument
            shift # past value
            ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
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

function configureMinikubeAddons() {
    # Enable metrics-server addon for kubectl top
    minikube addons enable metrics-server
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

    # In case apiserver is not available get minikube logs
    if [[ -z "$STATUS" ]] && [[ "$VM_DRIVER" = "none" ]]; then
      cat /var/lib/minikube/minikube.err
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
    local status=$(minikube status --format "{{.Host}}")
    if [ -n "${status}" ]; then
        log "Minikube is already initialized" red
        read -p "Do you want to remove previous minikube cluster [y/N]: " deleteMinikube
        if [ "${deleteMinikube}" == "y" ]; then
            minikube delete
        fi
    fi
}

function checkMinikubeVersion() {
    local version=$(minikube version | awk '{print  $3}' | grep -o '[0-9\.]\+' )
    local version_clean=$(echo $version | awk -F '.' '{print $1"."$2;}')
    local supported_version_min=$(echo ${MINIKUBE_VERSION} | awk -F '.' '{print $1"."--$2;}')
    local supported_version_max=$(echo ${MINIKUBE_VERSION} | awk -F '.' '{printf "%d.%d", $1, ++$2;}')

    if [[ "$(printf "${version_clean}\n${supported_version_min}" | sort -V | head -n1)" == "${version_clean}" ]]; then
        log "Your minikube is in ${version}. Your version is older than the supported version of minikube (v$MINIKUBE_VERSION)" yellow
    fi

    if [[ "$(printf "${version_clean}\n${supported_version_max}" | sort -V | head -n1)" == "${supported_version_max}" ]]; then
        log "Your minikube is in ${version}. Your version is newer than the supported version of minikube (v$MINIKUBE_VERSION)" yellow
    fi
}

function checkKubectlVersion() {
    local currentVersion=$(kubectl version --client --short | awk '{print $3}')
    local currentVersionMajor=$(echo ${currentVersion} | grep -o '[0-9]\+' | sed -n '1p')
    local currentVersionMinor=$(echo ${currentVersion} | grep -o '[0-9]\+' | sed -n '2p')
    local versionMajor=$(echo ${KUBECTL_CLI_VERSION} | cut -d"." -f1)
    local versionMinor=$(echo ${KUBECTL_CLI_VERSION} | cut -d"." -f2)
    local versionMinorDifference=$(( versionMinor - currentVersionMinor ))

    if [[ ${versionMinorDifference} -gt 1 ]] || [[ ${versionMinorDifference} -lt -1 ]]; then
        echo "Your kubectl version is ${currentVersion}. Supported versions of kubectl are from ${versionMajor}.$(( ${versionMinor} - 1 )).* to ${versionMajor}.$(( ${versionMinor} + 1 )).*"
    fi

    if [[ ${versionMajor} -ne ${currentVersionMajor} ]]; then
        echo "Your kubectl version is ${currentVersion}. Supported versions of kubectl are from ${versionMajor}.$(( ${versionMinor} - 1 )).* to ${versionMajor}.$(( ${versionMinor} + 1 )).*"
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
        # Filter out lines that DO NOT start with 127.0.0.1 AND contain the MINIKUBE_DOMAIN pattern
        awk -v domain="${MINIKUBE_DOMAIN}" '$1=="127.0.0.1"||!index($0,domain)' /etc/hosts > kyma-hosts-tmp
        cat kyma-hosts-tmp | sudo tee /etc/hosts > /dev/null
        rm kyma-hosts-tmp
    fi

    log "Adding ${hostnames} to /etc/hosts on localhost" yellow
	local hostAlias="${minikubeIP} ${hostnames}"

    #Set new host alias
    echo ${hostAlias} | sudo tee -a /etc/hosts > /dev/null

	log "Domain added to /etc/hosts" green
}

function increaseFsInotifyMaxUserInstances() {
    # Default value of 128 is not enough to perform “kubectl log -f” from pods, hence increased to 524288
    if [[ "$VM_DRIVER" != "none" ]]; then
        minikube ssh -- "sudo sysctl -w fs.inotify.max_user_instances=524288"
        log "fs.inotify.max_user_instances is increased" green
    fi
}

function applyDefaultRbacRole() {
    kubectl apply -f "${RESOURCES_DIR}/default-sa-rbac-role.yaml"
}

function start() {
    checkMinikubeVersion

    checkKubectlVersion

    checkIfMinikubeIsInitialized

    initializeMinikubeConfig

    if [[ "$VM_DRIVER" = "none" ]]; then
        fixDindMinikubeIssue
    fi

    minikube start \
    --memory $MEMORY \
    --cpus 4 \
    --extra-config=apiserver.authorization-mode=RBAC \
    --extra-config=apiserver.cors-allowed-origins="http://*" \
    --extra-config=apiserver.enable-admission-plugins="DefaultStorageClass,LimitRanger,MutatingAdmissionWebhook,NamespaceExists,NamespaceLifecycle,ResourceQuota,ServiceAccount,ValidatingAdmissionWebhook" \
    --kubernetes-version=v$KUBERNETES_VERSION \
    --vm-driver=$VM_DRIVER \
    --disk-size=$DISK_SIZE \
    --bootstrapper=kubeadm

    waitForMinikubeToBeUp

    configureMinikubeAddons

    # Adding domains to /etc/hosts files
    addDevDomainsToEtcHosts "apiserver.${MINIKUBE_DOMAIN} console.${MINIKUBE_DOMAIN} catalog.${MINIKUBE_DOMAIN} instances.${MINIKUBE_DOMAIN} brokers.${MINIKUBE_DOMAIN} dex.${MINIKUBE_DOMAIN} docs.${MINIKUBE_DOMAIN} addons.${MINIKUBE_DOMAIN} lambdas-ui.${MINIKUBE_DOMAIN} console-backend.${MINIKUBE_DOMAIN} minio.${MINIKUBE_DOMAIN} jaeger.${MINIKUBE_DOMAIN} grafana.${MINIKUBE_DOMAIN} log-ui.${MINIKUBE_DOMAIN} loki.${MINIKUBE_DOMAIN} configurations-generator.${MINIKUBE_DOMAIN} gateway.${MINIKUBE_DOMAIN} connector-service.${MINIKUBE_DOMAIN} oauth2.${MINIKUBE_DOMAIN} kiali.${MINIKUBE_DOMAIN} compass-gateway.${MINIKUBE_DOMAIN} compass-gateway-mtls.${MINIKUBE_DOMAIN} compass-gateway-auth-oauth.${MINIKUBE_DOMAIN} compass.${MINIKUBE_DOMAIN} compass-mf.${MINIKUBE_DOMAIN} core-ui.${MINIKUBE_DOMAIN}"

    increaseFsInotifyMaxUserInstances

    applyDefaultRbacRole
}

start
