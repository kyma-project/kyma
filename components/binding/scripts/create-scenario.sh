#!/bin/bash
set -o errexit

echo "Creating Scenario"

kubectl apply -f ./scripts/secret.yaml
kubectl apply -f ./scripts/deployment-sandbox.yaml
kubectl apply -f ./scripts/binding-sample.yaml