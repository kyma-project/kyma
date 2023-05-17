.PHONY: generate
generate: telemetry-docs cra-docs eventing-docs

.PHONY: telemetry-tracepipeline
telemetry-tracepipeline:
	go run main.go --crd-filename ../../installation/resources/crds/telemetry/tracepipelines.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-03-tracepipeline.md

.PHONY: telemetry-logpipeline
telemetry-logpipeline:
	go run main.go --crd-filename ../../installation/resources/crds/telemetry/logpipelines.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-01-logpipeline.md

.PHONY: telemetry-logparser
telemetry-logparser:
	go run main.go --crd-filename ../../installation/resources/crds/telemetry/logparsers.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-02-logparser.md

.PHONY: telemetry-docs
telemetry-docs: telemetry-tracepipeline telemetry-logparser telemetry-logpipeline

.PHONY: cra-docs
cra-docs:
	go run main.go --crd-filename ../../installation/resources/crds/compass-runtime-agent/compass-connection.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/ra-01-compassconnection.md

.PHONY: serverless-docs
serverless-docs:
	go run main.go --crd-filename ../../installation/resources/crds/serverless/crd-serverless.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/svls-01-function.md

.PHONY: eventing-subscription
eventing-subscription:
	go run main.go --crd-filename ../../installation/resources/crds/eventing/subscriptions.eventing.kyma-project.io.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/evnt-01-subscription.md

.PHONY: eventing-backend
eventing-backend:
	go run main.go --crd-filename ../../installation/resources/crds/eventing/eventingbackends.eventing.kyma-project.io.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/evnt-02-eventingbackend.md

.PHONY: eventing-docs
eventing-docs: eventing-subscription eventing-backend
