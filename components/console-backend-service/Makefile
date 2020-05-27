.PHONY: generate build build-dev

generate:
	go generate ./...

build:
	go build ./cmd/server

build-dev: generate	build

watch: export APP_VERBOSE=true
watch: export APP_KUBECONFIG_PATH=/Users/${USER}/.kube/config
watch: export APP_OIDC_ISSUER_URL=https://dex.${KYMA_DOMAIN}
watch: export APP_OIDC_CLIENT_ID=kyma-client

watch:
	compileDaemon -build="make build" -command="./server"
