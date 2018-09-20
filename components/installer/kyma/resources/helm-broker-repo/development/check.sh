#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

function show_help() {
    echo " Bundle checking tool"
    echo " Usage:"
    echo "   check.sh [flags] bundle"
    echo
    echo " Flags:"
    echo "    -h --help         helm for the script"
    echo "    -c --kube-context kube context to use (required if --dry-run selected)"
    echo "    --debug           prints output from helm command"
    echo "    --dry-run         performs helm install --dry-run operation"
    echo
    echo " Example of usage:"
    echo " ./check.sh -c minikube --dry-run redis-0.0.3"
}

silence="1> /dev/null"
dryRun=false

if [[ $# -eq 0 ]]
then
    show_help
    exit 0
fi

while [[ $# -gt 1 ]]
do
    case $1 in
        -h|-\?|--help)
            show_help
            exit 0
            ;;
        -c|--kube-context)
            if [ -n "$2" ]; then
                kubeContext=$2
                shift
            else
                echo "ERROR: '$1' requires a kube context\n" >&2
                exit 1
            fi
            ;;
        --debug)
            # no silencing
            silence=""
            ;;
        --dry-run)
            dryRun=true
            ;;
    esac
    shift
done


bundle=$1
if [ -z ${bundle} ];
then
    echo "ERROR: bundle not specified"
    exit 1
fi

if ! [ $(which checker) ]; then
    echo "Installing bundle checker"
    go get "github.com/kyma-project/helm-broker/cmd/checker"
    if [ $? -ne 0 ]
    then
        echo -e "${RED}Cannot install checker${NC}"
        exit 1
    fi
fi

# Bundle check
echo -e "${GREEN}"
checker "$1"
echo -e "${NC}"
if [ $? -ne 0 ]
then
    exit 1
fi

if [ "$dryRun" = false ];
then
    exit 0
fi

# -------------
# Helm dry run

if [ -z kubeContext ];
then
    echo -e "${RED}ERROR: kube context required${NC}"
    exit 1
fi

for chart in ${bundle}/chart/*/; do
    for plan in ${bundle}/plans/*/; do
        if [ -e ${plan}values.yaml ]
        then
            helmCmd="helm install --dry-run ${chart} --values ${plan}values.yaml --debug --kube-context ${kubeContext}"
        else
            helmCmd="helm install --dry-run ${chart} --debug --kube-context ${kubeContext}"
        fi
        echo -e "${GREEN}Executing: ${helmCmd}${NC}"
        eval "${helmCmd} ${silence}"
        if [ $? -eq 1 ];
        then
            echo -e "${RED}Could not perform helm install ${chart} with plan ${plan}"
            echo "Try to run command:"
            echo "${helmCmd}"
            echo -e "${NC}"
            exit 1
        fi
    done
done


