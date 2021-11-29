package serverless

import (
	"time"
)

type FunctionConfig struct {
	PublisherProxyAddress                       string        `envconfig:"default=http://eventing-publisher-proxy.kyma-system.svc.cluster.local/publish"`
	ImageRegistryDefaultDockerConfigSecretName  string        `envconfig:"default=serverless-registry-config-default"`
	ImageRegistryExternalDockerConfigSecretName string        `envconfig:"default=serverless-registry-config"`
	PackageRegistryConfigSecretName             string        `envconfig:"default=serverless-package-registry-config"`
	ImagePullAccountName                        string        `envconfig:"default=serverless-function"`
	BuildServiceAccountName                     string        `envconfig:"default=serverless-build"`
	TargetCPUUtilizationPercentage              int32         `envconfig:"default=50"`
	RequeueDuration                             time.Duration `envconfig:"default=1m"`
	FunctionReadyRequeueDuration                time.Duration `envconfig:"default=5m"`
	Build                                       BuildConfig
}

type GitConfig struct {
	GitFetchRequeueDuration time.Duration `envconfig:"default=30s"`
	Backoff                 GitBackoffConfig
}

type GitBackoffConfig struct {
	Duration time.Duration `envconfig:"default=10s"`
	Factor   float64       `envconfig:"default=2.0"`
	Jitter   float64       `envconfig:"default=0.0"`
	Steps    int           `envconfig:"default=5"`
	Cap      time.Duration `envconfig:"default=180s"`
}

type BuildConfig struct {
	ExecutorArgs        []string `envconfig:"default=--insecure;--skip-tls-verify;--skip-unused-stages;--log-format=text;--cache=true"`
	ExecutorImage       string   `envconfig:"default=gcr.io/kaniko-project/executor:v0.22.0"`
	RepoFetcherImage    string   `envconfig:"default=eu.gcr.io/kyma-project/function-build-init:305bee60"`
	MaxSimultaneousJobs int      `envconfig:"default=5"`
}

type DockerConfig struct {
	ActiveRegistryConfigSecretName string
	PushAddress                    string
	PullAddress                    string
}
