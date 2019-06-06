#!/bin/bash

set -o errexit

kubectl delete subscriptions.eventing.knative.dev --all -n kyma-system
kubectl delete channels.eventing.knative.dev --all -n kyma-system
kubectl apply -f ./crds.yaml
kubectl delete pod -l app=event-bus-subscription-controller-knative -n kyma-system