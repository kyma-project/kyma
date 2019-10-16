#!/bin/bash -e

# Create sample app & service
kubectl apply -f test-app.yaml
# Create apirule
kubectl apply -f apirule-test.yaml
# Connect to apirule
