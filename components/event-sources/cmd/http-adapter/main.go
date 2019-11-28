package main

import (
	"knative.dev/eventing/pkg/adapter"

	"github.com/kyma-project/kyma/components/event-sources/adapter/http"
)

func main() {
	adapter.Main("http-source", http.NewEnvConfig, http.NewAdapter)
}
