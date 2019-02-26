#!/usr/bin/env bash

SECONDS=0
END_TIME=$((SECONDS+60))

while [ ${SECONDS} -lt ${END_TIME} ];do
  EXTERNAL_PUBLIC_IP=$(kubectl get service -n ${SERVICE_NAMESPACE} ${SERVICE_NAME} -o jsonpath="{.status.loadBalancer.ingress[0].ip}")

  if [ "${EXTERNAL_PUBLIC_IP}" ]; then
      break
  fi

  sleep 10
  SECONDS=$((SECONDS+10))
done

if [ -z "${EXTERNAL_PUBLIC_IP}" ]; then
    (>&2 echo "External public IP not found")
    exit 1
fi

echo "$EXTERNAL_PUBLIC_IP"