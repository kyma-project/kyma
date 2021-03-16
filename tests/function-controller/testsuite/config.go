package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

const (
	TestDataKey  = "testData"
	EventPing    = "event-ping"
	RedisEnvPing = "env-ping"

	GotEventMsg              = "The event has come!"
	AnswerForEnvPing         = "Redis port: 6379"
	HappyMsg                 = "happy"
	AddonsConfigUrl          = "https://github.com/kyma-project/addons/releases/download/0.13.0/index-testing.yaml"
	ServiceClassExternalName = "redis"
	ServicePlanExternalName  = "micro"
	RedisEnvPrefix           = "REDIS_TEST_"
)

type Config struct {
	UsageKindName                   string           `envconfig:"default=serverless-function"`
	Namespace                       string           `envconfig:"default=test-function"`
	DomainName                      string           `envconfig:"default=kyma.local"`
	Verbose                         bool             `envconfig:"default=false"`
	WaitTimeout                     time.Duration    `envconfig:"default=15m"` // damn istio
	DomainPort                      uint32           `envconfig:"default=80"`
	MaxPollingTime                  time.Duration    `envconfig:"default=5m"`
	InsecureSkipVerify              bool             `envconfig:"default=true"`
	Cleanup                         step.CleanupMode `envconfig:"default=yes"`
	GitServerImage                  string           `envconfig:"default=eu.gcr.io/kyma-project/gitserver:PR-2696"`
	GitServerRepoName               string           `envconfig:"default=function"`
	IstioEnabled                    bool             `envconfig:"default=true"`
	PublishURL                      string           `envconfig:"default=http://eventing-event-publisher-proxy.kyma-system/publish"`
	PackageRegistryConfigSecretName string           `envconfig:"default=serverless-package-registry-config"`
	PackageRegistryConfigURLNode    string           `envconfig:"default=https://pkgs.dev.azure.com/kyma-wookiee/public-packages/_packaging/public-packages%40Release/npm/registry/"`
	PackageRegistryConfigURLPython  string           `envconfig:"default=https://pkgs.dev.azure.com/kyma-wookiee/public-packages/_packaging/public-packages%40Release/pypi/simple/"`
}
