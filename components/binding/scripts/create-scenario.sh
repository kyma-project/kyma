#!/bin/bash
set -o errexit

echo "Creating Scenario"

kubectl apply -f components/binding/scripts/secret.yaml
kubectl apply -f components/binding/scripts/deployment-sandbox.yaml
kubectl apply -f components/binding/scripts/binding-sample.yaml