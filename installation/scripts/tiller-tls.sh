#!/usr/bin/env bash -e

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

echo "---> Get Helm secrets and put then into $(helm home)"
kubectl get -n kyma-installer secret helm-secret -o json | jq '.data."global.helm.ca.crt"' > "$(helm home)/ca.pem"
kubectl get -n kyma-installer secret helm-secret -o json | jq '.data."global.helm.tls.crt"' > "$(helm home)/cert.pem"
kubectl get -n kyma-installer secret helm-secret -o json | jq '.data."global.helm.tls.key"' > "$(helm home)/key.pem"
