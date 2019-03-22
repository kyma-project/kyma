#!/usr/bin/env bash -e

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

while :
do
  if [[ $(kubectl get -n kyma-installer secret helm-secret) ]]
    then
      echo "---> Secrets have been created"
      break
    else
      echo "---> Secrets not present. Waiting 5s..."
      sleep 5
    fi
done

echo "---> Get Helm secrets and put then into $(helm home)"
kubectl get -n kyma-installer secret helm-secret -o json | jq '.data."global.helm.ca.crt"' > "$(helm home)/ca.pem"
kubectl get -n kyma-installer secret helm-secret -o json | jq '.data."global.helm.tls.crt"' > "$(helm home)/cert.pem"
kubectl get -n kyma-installer secret helm-secret -o json | jq '.data."global.helm.tls.key"' > "$(helm home)/key.pem"
