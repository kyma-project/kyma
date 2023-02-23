#!/usr/bin/env bash

kubectl delete crd podpresets.settings.svcat.k8s.io
kubectl delete mutatingwebhookconfiguration cluster-essentials-pod-preset-webhook
kubectl delete -nkyma-system secret cluster-essentials-pod-preset-webhook-cert
kubectl delete -nkyma-system serviceaccount cluster-essentials-pod-preset-webhook
kubectl delete clusterrole cluster-essentials-pod-preset-webhook
kubectl delete clusterrolebinding cluster-essentials-pod-preset-webhook
kubectl delete -nkyma-system service cluster-essentials-pod-preset-webhook
kubectl delete -nkyma-system deployment cluster-essentials-pod-preset-webhook
