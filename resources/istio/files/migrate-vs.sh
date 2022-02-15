set -e

mkdir -p istio/base
mkdir -p istio/merged

kubectl get vs -A -o custom-columns=NAME:.metadata.name,NAMESPACE:.metadata.namespace | tail -n +2 > istio/workload

cat << EOF > istio/cors-patch.yaml
---
spec:
  http:
  - match:
    corsPolicy:
      allowOrigins:
        - regex: ".*"
EOF

while read line; do
  NAME=$(echo "$line" | awk '{print $1}')
  NAMESPACE=$(echo "$line" | awk '{print $2}')

  kubectl get -n "${NAMESPACE}" vs "${NAME}" -o yaml  > istio/base/"${NAMESPACE}-${NAME}.yaml"
  if cat "istio/base/${NAMESPACE}-${NAME}.yaml" | grep -c "allowOrigin:"; then
  	yq merge -x "istio/base/${NAMESPACE}-${NAME}.yaml" "istio/cors-patch.yaml" > istio/merged/"${NAMESPACE}-${NAME}.yaml"
  fi

done < istio/workload

kubectl apply -f istio/merged/
