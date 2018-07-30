#!/bin/bash

set -o errexit

CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TPL_PATH="$CURRENT_DIR/../resources/installation-config.yaml.tpl"

EXTERNAL_IP_ADDRESS=""
DOMAIN=""
REMOTE_ENV_IP=""
K8S_APISERVER_URL=""
K8S_APISERVER_CA=""
ADMIN_GROUP=""
ENABLE_ETCD_BACKUP_OPERATOR=""
ETCD_BACKUP_ABS_CONTAINER_NAME=""

POSITIONAL=()
while [[ $# -gt 0 ]]
do
    key="$1"

    case ${key} in
        --ip-address)
            EXTERNAL_IP_ADDRESS="$2"
            shift # past argument
            shift # past value
            ;;
        --domain)
            DOMAIN="$2"
            shift # past argument
            shift # past value
            ;;
        --remote-env-ip)
            REMOTE_ENV_IP="$2"
            shift # past argument
            shift # past value
            ;;
        --k8s-apiserver-url)
            K8S_APISERVER_URL="$2"
            shift # past argument
            shift # past value
            ;;
        --k8s-apiserver-ca)
            K8S_APISERVER_CA="$2"
            shift # past argument
            shift # past value
            ;;
        --admin-group)
            ADMIN_GROUP="$2"
            shift # past argument
            shift # past value
            ;;
        --enable-etcd-backup-operator)
            ENABLE_ETCD_BACKUP_OPERATOR="$2"
            shift # past argument
            shift # past value
            ;;
        --etcd-backup-abs-container-name)
            ETCD_BACKUP_ABS_CONTAINER_NAME="$2"
            shift # past argument
            shift # past value
            ;;
        --output)
            OUTPUT="$2"
            shift
            shift
            ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

cp $TPL_PATH $OUTPUT

case `uname -s` in
    Darwin)
        sed -i "" "s;__EXTERNAL_IP_ADDRESS__;${EXTERNAL_IP_ADDRESS};" "$OUTPUT"
        sed -i "" "s;__DOMAIN__;${DOMAIN};" "$OUTPUT"
        sed -i "" "s;__REMOTE_ENV_IP__;${REMOTE_ENV_IP};" "$OUTPUT"
        sed -i "" "s;__K8S_APISERVER_URL__;${K8S_APISERVER_URL};" "$OUTPUT"
        sed -i "" "s;__K8S_APISERVER_CA__;${K8S_APISERVER_CA};" "$OUTPUT"
        sed -i "" "s;__ADMIN_GROUP__;${ADMIN_GROUP};" "$OUTPUT"
        sed -i "" "s;__ENABLE_ETCD_BACKUP_OPERATOR__;${ENABLE_ETCD_BACKUP_OPERATOR};" "$OUTPUT"
        sed -i "" "s;__ETCD_BACKUP_ABS_CONTAINER_NAME__;${ETCD_BACKUP_ABS_CONTAINER_NAME};" "$OUTPUT"
        ;;
    *)
        sed -i "s;__EXTERNAL_IP_ADDRESS__;${EXTERNAL_IP_ADDRESS};g" "$OUTPUT"
        sed -i "s;__DOMAIN__;${DOMAIN};g" "$OUTPUT"
        sed -i "s;__REMOTE_ENV_IP__;${REMOTE_ENV_IP};g" "$OUTPUT"
        sed -i "s;__K8S_APISERVER_URL__;${K8S_APISERVER_URL};g" "$OUTPUT"
        sed -i "s;__K8S_APISERVER_CA__;${K8S_APISERVER_CA};g" "$OUTPUT"
        sed -i "s;__ADMIN_GROUP__;${ADMIN_GROUP};g" "$OUTPUT"
        sed -i "s;__ENABLE_ETCD_BACKUP_OPERATOR__;${ENABLE_ETCD_BACKUP_OPERATOR};g" "$OUTPUT"
        sed -i "s;__ETCD_BACKUP_ABS_CONTAINER_NAME__;${ETCD_BACKUP_ABS_CONTAINER_NAME};g" "$OUTPUT"
        ;;
esac
