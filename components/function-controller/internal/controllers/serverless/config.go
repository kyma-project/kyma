package serverless

import (
	"time"
)

type FunctionConfig struct {
	PublisherProxyAddress                       string        `envconfig:"optional"`
	JaegerServiceEndpoint                       string        `envconfig:"optional"`
	TraceCollectorEndpoint                      string        `envconfig:"optional"`
	ImageRegistryDefaultDockerConfigSecretName  string        `envconfig:"default=serverless-registry-config-default"`
	ImageRegistryExternalDockerConfigSecretName string        `envconfig:"default=serverless-registry-config"`
	PackageRegistryConfigSecretName             string        `envconfig:"default=serverless-package-registry-config"`
	ImagePullAccountName                        string        `envconfig:"default=serverless-function"`
	TargetCPUUtilizationPercentage              int32         `envconfig:"default=50"`
	RequeueDuration                             time.Duration `envconfig:"default=1m"`
	FunctionReadyRequeueDuration                time.Duration `envconfig:"default=5m"`
	GitFetchRequeueDuration                     time.Duration `envconfig:"default=30s"`
	Build                                       BuildConfig
}

type BuildConfig struct {
	ExecutorArgs        []string `envconfig:"default=--insecure;--skip-tls-verify;--skip-unused-stages;--log-format=text;--cache=true"`
	ExecutorImage       string   `envconfig:"default=gcr.io/kaniko-project/executor:v0.22.0"`
	RepoFetcherImage    string   `envconfig:"default=europe-docker.pkg.dev/kyma-project/prod/function-build-init:v20230426-37b02524"`
	MaxSimultaneousJobs int      `envconfig:"default=5"`
}

type DockerConfig struct {
	ActiveRegistryConfigSecretName string
	PushAddress                    string
	PullAddress                    string
}
