set -ex
set -o pipefail

# # # # # # # # # # # # # # # # # # #
# VARs coming from environment:     #
#                                   #
# TYPE                              #
# APISERVER_SERVICE_NAME            #
# APISERVER_SERVICE_NAMESPACE       #
# # # # # # # # # # # # # # # # # # #

echo "Get required packages"
apk add gettext

getLoadBalancerIP() {

    if [ "$#" -ne 2 ]; then
        echo "usage: getLoadBalancerIP <service_name> <namespace>"
        exit 1
    fi

    local SERVICE_NAME="$1"
    local SERVICE_NAMESPACE="$2"
    local LOAD_BALANCER_IP=""

    SECONDS=0
    END_TIME=$((SECONDS+60))

    while [ ${SECONDS} -lt ${END_TIME} ];do

        LOAD_BALANCER_IP=$(kubectl get service -n "${SERVICE_NAMESPACE}" "${SERVICE_NAME}" -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

        if [ -n "${LOAD_BALANCER_IP}" ]; then
            break
        fi

        sleep 10

    done

    if [ -z "${LOAD_BALANCER_IP}" ]; then
        echo "---> Could not retrive the IP address. Verify if service ${SERVICE_NAME} exists in the namespace ${SERVICE_NAMESPACE}" >&2
        echo "---> Command executed: kubectl get service -n ${SERVICE_NAMESPACE} ${SERVICE_NAME} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'" >&2
        exit 1
    fi

    echo "${LOAD_BALANCER_IP}"
}

generateXipDomain() {

    if [ "$#" -ne 2 ]; then
        echo "usage: generateXipDomain <service_name> <namespace>"
        exit 1
    fi

    local SERVICE_NAME="$1"
    local SERVICE_NAMESPACE="$2"
    local EXTERNAL_PUBLIC_IP
    EXTERNAL_PUBLIC_IP=$(getLoadBalancerIP "${SERVICE_NAME}" "${SERVICE_NAMESPACE}")

    if [[ "$?" != 0 ]]; then
        echo "External public IP not found"
        exit 1
    fi

    echo "${EXTERNAL_PUBLIC_IP}.xip.io"
}

kubectl get secret -n cert-manager kyma-ca-key-pair -o jsonpath='{.data.tls\.crt}' | base64 --decode > /etc/ca-tls-cert/tls.crt

echo "Finding XIP domain for api-server LoadBalancer..."
export XIP_DOMAIN=$(generateXipDomain "${APISERVER_SERVICE_NAME}" "${APISERVER_SERVICE_NAMESPACE}")
echo "XIP domain for api-server LoadBalancer: ${XIP_DOMAIN}"

echo "Generating Certificate for ApiServer Ingress Gateway"

manifests=(
  certificate.yaml
)

for resource in "${manifests[@]}"; do
    envsubst <"/etc/manifests/$resource" | kubectl apply -f -
done

echo "Success."

