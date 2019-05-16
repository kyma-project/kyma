#!/usr/bin/env bash

#Helper script that creates overrides for manual installation with
#DNS domain and TLS certificates managed by the End User.

ERRMSG=""

if [[ -z "${DOMAIN}" ]]; then
  ERRMSG=" - DOMAIN not set\n"
fi

if [[ -z "${TLS_CERT}" ]]; then
	ERRMSG="${ERRMSG} - TLS_CERT not set\n"
fi

if [[ -z "${TLS_KEY}" ]]; then
	ERRMSG="${ERRMSG} - TLS_KEY not set\n"
fi

if [[ -n "${ERRMSG}" ]]; then
	printf "Error(s) occurred:\n"
  printf "${ERRMSG}\n"
  exit 1
fi

cat <<ENDOFMESSAGE

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: owndomain-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.domainName: "${DOMAIN}"
  global.tlsCrt: "${TLS_CERT}"
  global.tlsKey: "${TLS_KEY}"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: owndomain-knative-serving-overrides
  namespace: kyma-installer
  labels:
    installer: overrides
    component: knative-serving
    kyma-project.io/installation: ""
data:
  knative-serving.domainName: "${DOMAIN}"
---
ENDOFMESSAGE
