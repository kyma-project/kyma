package main

import "knative.dev/eventing/pkg/adapter"
import "github.com/kyma-project/kyma/components/event-sources/adapter/http"

func main() {
	adapter.Main("application-source", http.NewEnvConfig, http.NewAdapter)
}
