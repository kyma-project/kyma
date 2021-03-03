# This will delete the cluster and the docker registry
k3d cluster delete kyma
docker rm -f  registry.localhost
docker network rm kyma