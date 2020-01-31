
#!/usr/bin/bash -e -x

WORKING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [[ -z "${NAMESPACE}" ]]; then
  export NAMESPACE=istio-perf-test
fi

export CLUSTER_DOMAIN_NAME=$(kubectl get gateways.networking.istio.io kyma-gateway -n kyma-system -ojsonpath="{.spec.servers[0].hosts[0]}" | sed 's/*.//g' )
export WORKLOAD_SIZE=${WORKLOAD_SIZE:-10}
export VUS=${VUS:-64}

common_resources=(
	namespace.yaml
)

workload_resources=(
	app.yaml
	api.yaml
)

for resource in "${common_resources[@]}"; do
    envsubst <"${WORKING_DIR}/$resource" | kubectl -n "${NAMESPACE}" apply -f -
done

for (( i = 0; i < $WORKLOAD_SIZE; i++ )); do
	export WORKER=$(($i + 1))
	for resource in "${workload_resources[@]}"; do
    	envsubst <"${WORKING_DIR}/$resource" | kubectl -n "${NAMESPACE}" apply -f -

    	while :
        do
            if [ "$(kubectl get pod -l "app"="httpbin-$WORKER" -n ${NAMESPACE} -o jsonpath='{.items[0].status.containerStatuses[0].ready}')" = "true" ]; then
                echo "httpbin-$WORKER is running..."
                break
            else
                echo "waiting 3s..."
                sleep 3
            fi
        done

	done
done
