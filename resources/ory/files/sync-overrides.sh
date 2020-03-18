#!/bin/bash -e

# 1 - Get current secrets (from mounted secret)
# 2 - Get current overrides (kubectl get secret -n kyma-installer -l "installer=overrides,component=ory")
LOCAL_SECRETS=$(ls /etc/secrets)

# 1 - If override secret does not exist -> create
# 2 - patch secret with values

(cat << EOF
---
apiVersion: v1
kind: Secret
metadata:
  labels:
    installer: "overrides"
    component: "ory"
    generated: "true"
    kyma-project.io/installation: ""
  name: ory-overrides-generated
  namespace: kyma-installer
type: Opaque
EOF
) | kubectl apply -f -

for opt in $LOCAL_SECRETS
do
	PATCH=$(cat << EOF
---
data:
  $(echo ${opt}: $(base64 /etc/secrets/${opt} | sed 's/ /\\ /g' | tr -d '\n'))
EOF
)
    set +e
    msg=$(kubectl patch secret ory-overrides-generated --patch "${PATCH}" -n kyma-installer 2>&1)
    status=$?
    set -e
    if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
        echo "$msg"
        exit $status
    fi
done
