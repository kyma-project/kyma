#!/usr/bin/env bash

kubectl create ns serverless || true
kubectl apply -f https://raw.githubusercontent.com/kubeless/kubeless/master/examples/nodejs/function1.yaml -n serverless
kubectl create ns python || true
kubectl apply -f https://raw.githubusercontent.com/kubeless/kubeless/master/examples/python/function1.yaml -n python
kubectl create ns testing-lambdas-hard || true
kubectl apply -f crdToApply -n testing-lambdas-hard