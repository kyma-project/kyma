.PHONY: generate
generate: cra-docs eventing-docs apix-docs

.PHONY: cra-docs
cra-docs:
	go run main.go --crd-filename ../../installation/resources/crds/compass-runtime-agent/compass-connection.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/ra-01-compassconnection.md

.PHONY: serverless-docs
serverless-docs:
	go run main.go --crd-filename ../../installation/resources/crds/serverless/crd-serverless.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/svls-01-function.md --ignore-spec env.valueFrom --ignore-spec env.name --ignore-spec env.value --ignore-spec resourceConfiguration.build.resources.requests --ignore-spec resourceConfiguration.build.resources.limits --ignore-spec resourceConfiguration.function.resources.requests --ignore-spec resourceConfiguration.function.resources.limits

.PHONY: eventing-subscription
eventing-subscription:
	go run main.go --crd-filename ../../installation/resources/crds/eventing/subscriptions.eventing.kyma-project.io.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/evnt-01-subscription.md

.PHONY: eventing-backend
eventing-backend:
	go run main.go --crd-filename ../../installation/resources/crds/eventing/eventingbackends.eventing.kyma-project.io.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/evnt-02-eventingbackend.md

.PHONY: eventing-docs
eventing-docs: eventing-subscription eventing-backend

.PHONY: apix-docs
apix-docs:
	go run main.go --crd-filename ../../installation/resources/crds/api-gateway/apirules.gateway.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/apix-01-apirule.md