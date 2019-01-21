#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCES_DIR="${CURRENT_DIR}/../resources"

MINIKUBE_DOMAIN=""
MINIKUBE_VERSION=0.28.2
KUBERNETES_VERSION=1.10.0
KUBECTL_CLI_VERSION=1.10.0
VM_DRIVER=hyperkit
DISK_SIZE=20g
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
    # Enable heapster addon
    minikube addons enable heapster
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
        fi
    fi
}

function checkMinikubeVersion() {
    local version=$(minikube version | awk '{print  $3}')

    if [[ "${version}" != *"${MINIKUBE_VERSION}"* ]]; then
        echo "Your minikube is in ${version}. v${MINIKUBE_VERSION} is supported version of minikube. Install supported version!"
        exit -1
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
    --extra-config=apiserver.Authorization.Mode=RBAC \
    --extra-config=apiserver.GenericServerRunOptions.CorsAllowedOriginList=".*" \
    --extra-config=controller-manager.ClusterSigningCertFile="/var/lib/localkube/certs/ca.crt" \
    --extra-config=controller-manager.ClusterSigningKeyFile="/var/lib/localkube/certs/ca.key" \
    --extra-config=apiserver.admission-control="LimitRanger,ServiceAccount,DefaultStorageClass,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota" \
    --kubernetes-version=v$KUBERNETES_VERSION \
    --vm-driver=$VM_DRIVER \
    --disk-size=$DISK_SIZE \
    --feature-gates="MountPropagation=false" \
    -b=localkube

    waitForMinikubeToBeUp

    # Adding domains to /etc/hosts files
    addDevDomainsToEtcHosts "apiserver.${MINIKUBE_DOMAIN} console.${MINIKUBE_DOMAIN} catalog.${MINIKUBE_DOMAIN} instances.${MINIKUBE_DOMAIN} brokers.${MINIKUBE_DOMAIN} dex.${MINIKUBE_DOMAIN} docs.${MINIKUBE_DOMAIN} lambdas-ui.${MINIKUBE_DOMAIN} ui-api.${MINIKUBE_DOMAIN} minio.${MINIKUBE_DOMAIN} jaeger.${MINIKUBE_DOMAIN} grafana.${MINIKUBE_DOMAIN}  configurations-generator.${MINIKUBE_DOMAIN} gateway.${MINIKUBE_DOMAIN} connector-service.${MINIKUBE_DOMAIN}"

    increaseFsInotifyMaxUserInstances

    applyDefaultRbacRole
}

start
