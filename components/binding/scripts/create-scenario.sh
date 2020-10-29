#!/bin/bash
set -o errexit

echo "Creating Scenario"

pwd
kubectl apply -f ./secret.yaml
kubectl apply -f ./deployment-sandbox.yaml
kubectl apply -f ./binding-sample.yaml