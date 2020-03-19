#!/bin/bash -e
LOCAL_SECRETS=$(ls /etc/secrets)

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
  namespace: kyma-installer
type: Opaque
EOF
) | kubectl apply -f -

for key in $LOCAL_SECRETS
do
	value=$(base64 /etc/secrets/${key} | sed 's/ /\\ /g' | tr -d '\n')
	case $key in
		dsn* )
			override_key="hydra.config.dsn"
			;;
		secretsSystem* )
			override_key="hydra.config.secrets.system"
			;;
		secretsCookie* )
			override_key="hydra.config.secrets.cookie"
			;;
		postgresql-password* | dbPassword* )
			override_key="global.ory.hydra.persistence.password"
			;;
		* )
			override_key=$key
			;;
	esac

	PATCH=$(cat << EOF
---
data:
  $(echo ${override_key}: ${value})
EOF
)
    set +e
    msg=$(kubectl patch secret "{{ template "ory.fullname" . }}-overrides-generated" --patch "${PATCH}" -n kyma-installer 2>&1)
    status=$?
    set -e
    if [[ $status -ne 0 ]] && [[ ! "$msg" == *"not patched"* ]]; then
        echo "$msg"
        exit $status
    fi
done
