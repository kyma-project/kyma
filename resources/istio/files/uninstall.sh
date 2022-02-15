#!/bin/bash -e

istioctl x uninstall --purge -y
kubectl delete cm "${CONFIGMAP_NAME}" -n "${NAMESPACE}"
