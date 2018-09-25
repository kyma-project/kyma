# Local Kyma installation

./installation/scripts/minikube.sh --domain "kyma.local" --vm-driver "hyperkit"

kc apply -f ./installation/resources/default-sa-rbac-role.yaml

wait for kube-dns running 3/3

./installation/scripts/install-tiller.sh

kc apply -f ./release/local-kyma-installer.yaml

kubectl label installation/kyma-installation action=install

./installation/scripts/is-installed.sh
