package internal

import (
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"time"
)

const (
	TestDataKey = "testData"
)

type Config struct {
	Namespace                       string               `envconfig:"default=test-function"`
	KubectlProxyEnabled             bool                 `envconfig:"default=false"`
	Verbose                         bool                 `envconfig:"default=false"`
	WaitTimeout                     time.Duration        `envconfig:"default=15m"`
	MaxPollingTime                  time.Duration        `envconfig:"default=5m"`
	InsecureSkipVerify              bool                 `envconfig:"default=true"`
	Cleanup                         executor.CleanupMode `envconfig:"default=yes"`
	GitServerImage                  string               `envconfig:"default=eu.gcr.io/kyma-project/gitserver:b60c3054"`
	GitServerRepoName               string               `envconfig:"default=function"`
	IstioEnabled                    bool                 `envconfig:"default=false"`
	PackageRegistryConfigSecretName string               `envconfig:"default=serverless-package-registry-config"`
	PackageRegistryConfigURLNode    string               `envconfig:"default=https://pkgs.dev.azure.com/kyma-wookiee/public-packages/_packaging/public-packages%40Release/npm/registry/"`
	PackageRegistryConfigURLPython  string               `envconfig:"default=https://pkgs.dev.azure.com/kyma-wookiee/public-packages/_packaging/public-packages%40Release/pypi/simple/"`
}
