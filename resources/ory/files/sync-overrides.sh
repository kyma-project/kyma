#!/bin/bash -e
LOCAL_SECRETS=$(ls /etc/secrets)
OVERRIDES_NAMESPACE="kyma-installer"
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
  name: {{ template "ory.fullname" . }}-overrides-generated
  namespace: ${OVERRIDES_NAMESPACE}
type: Opaque
EOF
) | kubectl apply -f -

for key in $LOCAL_SECRETS
do
	case $key in
		secretsSystem* )
			override_key="hydra.hydra.config.secrets.system"
			;;
		secretsCookie* )
			override_key="hydra.hydra.config.secrets.cookie"
			;;
		postgresql-password* | dbPassword* )
			override_key="global.ory.hydra.persistence.password"
			;;
    gcp-sa.json* )
      override_key="global.ory.hydra.persistence.gcloud.saJson"
      ;;
		* )
			continue
			;;
	esac

	PATCH=$(cat << EOF
---
stringData:
  $(echo ${override_key}: $(cat /etc/secrets/${key}))
EOF
)
    set +e
    msg=$(kubectl patch secret "{{ template "ory.fullname" . }}-overrides-generated" --patch "${PATCH}" -n "${OVERRIDES_NAMESPACE}" 2>&1)
    status=$?
    set -e
    if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
        echo "$msg"
        exit $status
    fi
done
