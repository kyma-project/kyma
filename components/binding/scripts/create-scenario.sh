#!/bin/bash
set -o errexit

echo "Creating Scenario"

kubectl apply -f ./scripts/scenario/secret.yaml
kubectl apply -f ./scripts/scenario/deployment.yaml
kubectl apply -f ./scripts/scenario/target-kind.yaml
kubectl apply -f ./scripts/scenario/binding.yaml