
#!/usr/bin/bash -e -x

WORKING_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export NAMESPACE=istio-perf-test
export CLUSTER_DOMAIN=$(kubectl get gateways.networking.istio.io kyma-gateway -n kyma-system -ojsonpath="{.spec.servers[0].hosts[0]}" | sed 's/*.//g' )
export WORKLOAD_SIZE=${WORKLOAD_SIZE:-10}

common_resources=(
	namespace.yaml
)

workload_resources=(
	app.yaml
	api.yaml
)

for resource in "${common_resources[@]}"; do
    echo "---> Deploying: $resource"
    envsubst <"${WORKING_DIR}/$resource" | kubectl -n "${NAMESPACE}" apply -f -
done

for (( i = 0; i < $WORKLOAD_SIZE; i++ )); do
	export WORKER=$(($i + 1))
	echo "---> Workload Count: ${WORKER}"
	for resource in "${workload_resources[@]}"; do
    	echo "------> Deploying: $resource"
    	envsubst <"${WORKING_DIR}/$resource" | kubectl -n "${NAMESPACE}" apply -f -
	done
done

sleep 3s

echo "---> DOMAIN: ${CLUSTER_DOMAIN}"