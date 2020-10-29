#!/bin/bash
set -o errexit

echo "Creating Scenario"

pwd
kubectl apply -f ./scripts/secret.yaml
kubectl apply -f ./scripts/deployment-sandbox.yaml
kubectl apply -f ./scripts/binding-sample.yaml