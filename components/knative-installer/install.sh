#!/usr/bin/env bash
######### Knative build & serving ########

echo "Installing Knative build and serving ..."

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

sed 's/LoadBalancer/NodePort/' </serving.yaml \
| tee knative-serving.yaml \
| kubectl apply -f -


echo "Verifying Knative build and serving installation..."
sleep 2
until kubectl get -f knative-serving.yaml > /dev/null 2>&1
do
    echo "Knative CRDs not yet synced, re-applying..."
    kubectl apply -f knative-serving.yaml
    sleep 2
done

${DIR}/is-ready.sh knative-build app build-controller
${DIR}/is-ready.sh knative-build app build-webhook
${DIR}/is-ready.sh knative-serving app activator
${DIR}/is-ready.sh knative-serving app autoscaler
${DIR}/is-ready.sh knative-serving app controller
${DIR}/is-ready.sh knative-serving app webhook

echo "Knative build and serving installation verified"

echo "Installing Knative eventing..."

kubectl apply -f /eventing.yaml

echo "Verifying Knative eventing installation..."

${DIR}/is-ready.sh knative-eventing app eventing-controller
${DIR}/is-ready.sh knative-eventing app webhook

echo "Knative eventing installation verified"
