package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

const (
	TestDataKey = "testData"
)

type Config struct {
	Namespace                       string           `envconfig:"default=test-function"`
	DomainName                      string           `envconfig:"default=kyma.local"`
	Verbose                         bool             `envconfig:"default=false"`
	WaitTimeout                     time.Duration    `envconfig:"default=15m"`
	DomainPort                      uint32           `envconfig:"default=80"`
	MaxPollingTime                  time.Duration    `envconfig:"default=5m"`
	InsecureSkipVerify              bool             `envconfig:"default=true"`
	Cleanup                         step.CleanupMode `envconfig:"default=yes"`
	GitServerImage                  string           `envconfig:"default=eu.gcr.io/kyma-project/gitserver:470796a1"`
	GitServerRepoName               string           `envconfig:"default=function"`
	IstioEnabled                    bool             `envconfig:"default=true"`
	PublishURL                      string           `envconfig:"default=http://eventing-event-publisher-proxy.kyma-system/publish"`
	PackageRegistryConfigSecretName string           `envconfig:"default=serverless-package-registry-config"`
	PackageRegistryConfigURLNode    string           `envconfig:"default=https://pkgs.dev.azure.com/kyma-wookiee/public-packages/_packaging/public-packages%40Release/npm/registry/"`
	PackageRegistryConfigURLPython  string           `envconfig:"default=https://pkgs.dev.azure.com/kyma-wookiee/public-packages/_packaging/public-packages%40Release/pypi/simple/"`
}
