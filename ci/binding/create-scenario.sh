#!/bin/bash
set -o errexit

echo "Creating Scenario"

kubectl apply -f ci/binding/secret.yaml
kubectl apply -f ci/binding/deployment-sandbox.yaml
kubectl apply -f ci/binding/binding-sample.yaml