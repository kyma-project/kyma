package serverless

import (
	"time"
)

type FunctionConfig struct {
	ImageRegistryDockerConfigSecretName string        `envconfig:"default=serverless-image-pull-secret"`
	ImagePullAccountName                string        `envconfig:"default=serverless"`
	TargetCPUUtilizationPercentage      int32         `envconfig:"default=50"`
	RequeueDuration                     time.Duration `envconfig:"default=1m"`
	GitFetchRequeueDuration             time.Duration `envconfig:"default=30s"`
	MaxConcurrentReconciles             int           `envconfig:"default=1"`
	Build                               BuildConfig
	Docker                              DockerConfig
}

type BuildConfig struct {
	ExecutorArgs     []string `envconfig:"default=--insecure;--skip-tls-verify;--skip-unused-stages;--log-format=text;--cache=true"`
	ExecutorImage    string   `envconfig:"default=gcr.io/kaniko-project/executor:v0.22.0"`
	RepoFetcherImage string   `envconfig:"default=eu.gcr.io/kyma-project/function-build-init:305bee60"`
}

type DockerConfig struct {
	InternalRegistryEnabled bool   `envconfig:"default=true"`
	InternalServerAddress   string `envconfig:"default=serverless-docker-registry.kyma-system.svc.cluster.local:5000"`
	RegistryAddress         string `envconfig:"default=registry.kyma.local"`
}
