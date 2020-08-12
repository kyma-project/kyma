package serverless

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

type FunctionConfig struct {
	ImageRegistryDockerConfigSecretName string        `envconfig:"default=serverless-image-pull-secret"`
	ImagePullAccountName                string        `envconfig:"default=serverless"`
	TargetCPUUtilizationPercentage      int32         `envconfig:"default=50"`
	RequeueDuration                     time.Duration `envconfig:"default=1m"`
	Build                               BuildConfig
	Docker                              DockerConfig
}

type BuildConfig struct {
	RequestsCPU          string            `envconfig:"default=350m"`
	RequestsCPUValue     resource.Quantity `envconfig:"-"`
	RequestsMemory       string            `envconfig:"default=750Mi"`
	RequestsMemoryValue  resource.Quantity `envconfig:"-"`
	LimitsCPU            string            `envconfig:"default=1"`
	LimitsCPUValue       resource.Quantity `envconfig:"-"`
	LimitsMemory         string            `envconfig:"default=1Gi"`
	LimitsMemoryValue    resource.Quantity `envconfig:"-"`
	RuntimeConfigMapName string            `envconfig:"default=dockerfile-nodejs-12"`
	ExecutorArgs         []string          `envconfig:"default=--insecure;--skip-tls-verify;--skip-unused-stages;--log-format=text;--cache=true"`
	ExecutorImage        string            `envconfig:"default=gcr.io/kaniko-project/executor:v0.22.0"`
	RepoFetcherImage     string            `envconfig:"eu.gcr.io/kyma-project/function-build-init:305bee60"`
}

type DockerConfig struct {
	InternalRegistryEnabled bool   `envconfig:"default=true"`
	InternalServerAddress   string `envconfig:"default=serverless-docker-registry.kyma-system.svc.cluster.local:5000"`
	RegistryAddress         string `envconfig:"default=registry.kyma.local"`
}
