#!/usr/bin/env sh
set -e

CURRENT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"

NAMESPACE=${NAMESPACE:-"kyma-system"}

# ETCD_CLUSTER_NAME - should be set as value of EtcdCluster CR's name

function generateCertificate {
    certProfile=$1
    cfssl gencert -ca=${tempPath}/ca.pem -ca-key=${tempPath}/ca-key.pem -config=${configPath}/ca-config.json -profile=${certProfile} ${configPath}/${certProfile}.json | cfssljson -bare ${tempPath}/${certProfile}
}

caSecretName="${ETCD_CLUSTER_NAME}-etcd-ca-tls"
serverSecretName="${ETCD_CLUSTER_NAME}-etcd-server-tls"
peerSecretName="${ETCD_CLUSTER_NAME}-etcd-peer-tls"
clientSecretName="${ETCD_CLUSTER_NAME}-etcd-client-tls"

tempPath="/tmp"
configPath="${CURRENT_DIR}/../config"

sed "s#__ETCD_CLUSTER_NAME__#$ETCD_CLUSTER_NAME#g
    s#__NAMESPACE__#$NAMESPACE#g
    " ${configPath}/peer.json.tpl > ${configPath}/peer.json
sed "s#__ETCD_CLUSTER_NAME__#$ETCD_CLUSTER_NAME#g
    s#__NAMESPACE__#$NAMESPACE#g
    " ${configPath}/server.json.tpl > ${configPath}/server.json

if [[ "$(kubectl get secret ${caSecretName} -n ${NAMESPACE})" ]]; then
    echo "CA certificate found. Downloading..."
    kubectl get secret ${caSecretName} -n ${NAMESPACE} -o=jsonpath='{.data.ca\.pem}' | base64 -d > ${tempPath}/ca.pem
    kubectl get secret ${caSecretName} -n ${NAMESPACE} -o=jsonpath='{.data.ca-key\.pem}' | base64 -d > ${tempPath}/ca-key.pem
else
    echo "CA certificate does not exist. Generating new one..."
    cfssl gencert -initca ${configPath}/ca-csr.json | cfssljson -bare ${tempPath}/ca -
    kubectl create secret generic ${caSecretName} --from-file=${tempPath}/ca.pem --from-file=${tempPath}/ca-key.pem --dry-run -o yaml > ${tempPath}/ca-secret.yaml
    kubectl apply -f ${tempPath}/ca-secret.yaml
fi


echo "generating etcd peer certs ==="
generateCertificate peer

echo "generating etcd server certs ==="
generateCertificate server

echo "generating etcd client certs ==="
generateCertificate client

mv ${tempPath}/client.pem ${tempPath}/etcd-client.crt
mv ${tempPath}/client-key.pem ${tempPath}/etcd-client.key
cp ${tempPath}/ca.pem ${tempPath}/etcd-client-ca.crt

mv ${tempPath}/server.pem ${tempPath}/server.crt
mv ${tempPath}/server-key.pem ${tempPath}/server.key
cp ${tempPath}/ca.pem ${tempPath}/server-ca.crt

mv ${tempPath}/peer.pem ${tempPath}/peer.crt
mv ${tempPath}/peer-key.pem ${tempPath}/peer.key
cp ${tempPath}/ca.pem ${tempPath}/peer-ca.crt

kubectl create secret generic ${peerSecretName} --from-file=${tempPath}/peer-ca.crt --from-file=${tempPath}/peer.crt --from-file=${tempPath}/peer.key --dry-run -o yaml > ${tempPath}/peers-secret.yaml
kubectl create secret generic ${serverSecretName} --from-file=${tempPath}/server-ca.crt --from-file=${tempPath}/server.crt --from-file=${tempPath}/server.key --dry-run -o yaml > ${tempPath}/server-secret.yaml
kubectl create secret generic ${clientSecretName} --from-file=${tempPath}/etcd-client-ca.crt --from-file=${tempPath}/etcd-client.crt --from-file=${tempPath}/etcd-client.key --dry-run -o yaml > ${tempPath}/client-secret.yaml

kubectl apply -f ${tempPath}/peers-secret.yaml
kubectl apply -f ${tempPath}/server-secret.yaml
kubectl apply -f ${tempPath}/client-secret.yaml