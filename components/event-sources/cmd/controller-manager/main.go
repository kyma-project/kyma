package main

import (
	"os"

	pkgerrors "github.com/pkg/errors"

	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/system"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	// allow client authentication against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/kyma-project/kyma/components/event-sources/reconciler/httpsource"
)

const (
	defaultSystemNamespace = "kyma-system"
	defaultMetricsDomain   = "kyma-project.io/event-sources"
)

func init() {
	if err := ensureEnvVars(); err != nil {
		panic(pkgerrors.Wrap(err, "setting environment variables"))
	}
}

func main() {
	sharedmain.Main("event_sources_controller",
		httpsource.NewController,
	)
}

func ensureEnvVars() error {
	envVarDefaults := map[string]string{
		system.NamespaceEnvKey: defaultSystemNamespace,
		metrics.DomainEnv:      defaultMetricsDomain,
	}

	var errs []error
	for ev, def := range envVarDefaults {
		if err := ensureEnvVar(ev, def); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

func ensureEnvVar(key, defaultVal string) error {
	if ns := os.Getenv(key); ns == "" {
		return os.Setenv(key, defaultVal)
	}
	return nil
}
