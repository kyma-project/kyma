# Support for old way of managing certificates for Minikube and Prow only
echo "${GLOBAL_TLS_KEY}" | base64 -d > ${HOME}/key.pem
echo "${GLOBAL_TLS_CERT}" | base64 -d > ${HOME}/cert.pem
kubectl create secret tls kyma-gateway-certs -n istio-system --key ${HOME}/key.pem --cert ${HOME}/cert.pem
