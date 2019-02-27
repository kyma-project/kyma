#!/usr/bin/env bash

#This script is only required for migration from 0.7 to 0.8

#The script is used to delete a LoadBalancer created by application-connector in version 0.7 of Kyma.
#In Kyma version 0.8 this LoadBalancer is extracted to a application-connector-xip chart to enable xip.io-related functionality.
#It means that during upgrade from 0.7 to 0.8 there is a time when old service is still running and a new one is created.
#The old service must be deleted to allow the new one to start, because the services are bound to the same IP Address.

#Hardcoded on purpose, this is not meant to be reusable
SERVICE_TO_REMOVE="application-connector-nginx-ingress-controller"
NAMESPACE="kyma-integration"

SERVICE_EXISTS=$(kubectl get service "${SERVICE_TO_REMOVE}" -n "${NAMESPACE}")

if [ -z "${SERVICE_EXISTS}" ]; then
  echo "Deleting service: ${NAMESPACE}/${SERVICE_TO_REMOVE}"
  kubectl delete service "${SERVICE_TO_REMOVE}" -n "${NAMESPACE}"
fi
