#!/usr/bin/env bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

DOMAIN="kyma.local"

VM_DRIVER="virtualbox"
if [ `uname -s` = "Darwin" ]; then
    VM_DRIVER="hyperkit"
fi

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --skip-minikube-start)
            SKIP_MINIKUBE_START=true
            shift # past argument
        ;;
        --cr)
            CR_PATH="$2"
            shift # past argument
            shift # past value
        ;;
        --crtpl_path)
            CRTPL_PATH="$2"
            shift
            shift
        ;;
        --vm-driver)
            VM_DRIVER="$2"
            shift
            shift
        ;;
        --serving-url)
            KN_SERVING_URL="$2"
            shift
            shift
        ;;
        --eventing-url)
            KN_EVENTING_URL="$2"
            shift
            shift
        ;;
        --knative)
            KNATIVE=1
            shift
        ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
        ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

MINIKUBE_ARGS="--domain \"${DOMAIN}\" --vm-driver \"${VM_DRIVER}\""

if [[ $KNATIVE ]]; then
    MINIKUBE_ARGS="${MINIKUBE_ARGS} --kubeadm --disk-size 30g"
fi

if [[ ! ${SKIP_MINIKUBE_START} ]]; then
    bash ${CURRENT_DIR}/../scripts/minikube.sh $MINIKUBE_ARGS
fi

if [[ $KNATIVE ]]; then
    if [ -n "${KN_SERVING_URL}" ]
    then
        EXTRA_ARGS="--serving-url ${KN_SERVING_URL}"
    fi

    if [ -n "${KN_EVENTING_URL}" ]
    then
        EXTRA_ARGS="${EXTRA_ARGS} --eventing-url ${KN_EVENTING_URL}"
    fi

    bash ${CURRENT_DIR}/../scripts/knative-install.sh ${EXTRA_ARGS}
    bash ${CURRENT_DIR}/../scripts/build-kyma-installer.sh --vm-driver "${VM_DRIVER}"

    bash ${CURRENT_DIR}/../scripts/generate-local-config.sh

    CRTPL_PATH=${CRTPL_PATH:-"$CURRENT_DIR/../resources/installer-cr-knative.yaml.tpl"}

    if [ -z "$CR_PATH" ]; then

        TMPDIR=`mktemp -d "${CURRENT_DIR}/../../temp-XXXXXXXXXX"`
        CR_PATH="${TMPDIR}/installer-cr-local.yaml"
        bash ${CURRENT_DIR}/../scripts/create-cr.sh --output "${CR_PATH}" --domain "${DOMAIN}" --crtpl_path "${CRTPL_PATH}"

    fi



    bash ${CURRENT_DIR}/../scripts/installer.sh --local --knative --cr "${CR_PATH}"
    rm -rf $TMPDIR
else
    bash ${CURRENT_DIR}/../scripts/build-kyma-installer.sh --vm-driver "${VM_DRIVER}"

    bash ${CURRENT_DIR}/../scripts/generate-local-config.sh

    CRTPL_PATH=${CRTPL_PATH:-"$CURRENT_DIR/../resources/installer-cr.yaml.tpl"}

    if [ -z "$CR_PATH" ]; then

        TMPDIR=`mktemp -d "${CURRENT_DIR}/../../temp-XXXXXXXXXX"`
        CR_PATH="${TMPDIR}/installer-cr-local.yaml"
        bash ${CURRENT_DIR}/../scripts/create-cr.sh --output "${CR_PATH}" --domain "${DOMAIN}" --crtpl_path "${CRTPL_PATH}"

    fi

    bash ${CURRENT_DIR}/../scripts/installer.sh --local --cr "${CR_PATH}"
    rm -rf $TMPDIR
fi