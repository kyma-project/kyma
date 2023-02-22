#!/usr/bin/env bash

echo "Deleting the compass-system namespace"
kubectl delete namespace compass-system
echo "Refreshing credential for Compass Connection"
kubectl patch compassconnection compass-connection --type=merge -p '{"spec":{"refreshCredentialsNow":true}}'