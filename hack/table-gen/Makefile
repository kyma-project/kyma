.PHONY: generate
generate: cra-docs apix-docs

.PHONY: cra-docs
cra-docs:
	go run main.go --crd-filename ../../installation/resources/crds/compass-runtime-agent/compass-connection.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/ra-01-compassconnection.md

.PHONY: apix-docs
apix-docs:
	go run main.go --crd-filename ../../installation/resources/crds/api-gateway/apirules.gateway.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/apix-01-apirule.md
