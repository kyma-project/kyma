.PHONY: telemetry-tracepipeline
telemetry-tracepipeline:
	go run table-gen.go --crd-filename ../../installation/resources/crds/telemetry/tracepipelines.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-03-tracepipeline.md

.PHONY: telemetry-logpipeline
telemetry-logpipeline:
	go run table-gen.go --crd-filename ../../installation/resources/crds/telemetry/logpipelines.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-01-logpipeline.md

.PHONY: telemetry-logparser
telemetry-logparser:
	go run table-gen.go --crd-filename ../../installation/resources/crds/telemetry/logparsers.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-02-logparser.md

.PHONY: telemetry-docs
telemetry-docs: telemetry-tracepipeline telemetry-logparser telemetry-logpipeline

.PHONY: application-connector-docs
application-connector-docs:
	go run table-gen.go --crd-filename ../../installation/resources/crds/application-connector/applications.applicationconnector.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/ac-01-application.md

.PHONY: generate
generate: telemetry-docs application-connector-docs
