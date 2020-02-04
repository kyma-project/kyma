/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
		httpsource.NewControllerWrapper,
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
