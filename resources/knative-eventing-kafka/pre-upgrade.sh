#!/usr/bin/env bash

set -eux -o pipefail

kubectl get deployment -n knative-eventing -lapp.kubernetes.io/instance=knative-eventing-kafka --ignore-not-found
kubectl delete deployment -n knative-eventing -lapp.kubernetes.io/instance=knative-eventing-kafka --ignore-not-found
