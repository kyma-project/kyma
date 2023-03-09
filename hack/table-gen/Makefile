tracepipeline:
	go run table-gen.go --crd-filename ../../installation/resources/crds/telemetry/tracepipelines.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-03-tracepipeline.md --crd-title TracePipeline

logpipeline:
	go run table-gen.go --crd-filename ../../installation/resources/crds/telemetry/logpipelines.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-01-logpipeline.md --crd-title LogPipeline

logparser:
	go run table-gen.go --crd-filename ../../installation/resources/crds/telemetry/logparsers.crd.yaml --md-filename ../../docs/05-technical-reference/00-custom-resources/telemetry-02-logparser.md --crd-title LogParser

.PHONY: run
run: tracepipeline logparser logpipeline
