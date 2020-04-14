package serverless

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

type FunctionConfig struct {
	ImagePullSecretName  string        `envconfig:"default=function-controller-registry-credentials"`
	ImagePullAccountName string        `envconfig:"default=function-controller"`
	RequeueDuration      time.Duration `envconfig:"default=1m"`
	Build                BuildConfig
	Docker               DockerConfig
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
	RuntimeConfigMapName string            `envconfig:"default=dockerfile-nodejs-8"`
	ExecutorImage        string            `envconfig:"default=gcr.io/kaniko-project/executor:v0.17.1"`
	CredsInitImage       string            `envconfig:"default=gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/creds-init:v0.10.1"`
}

type DockerConfig struct {
	Address         string `envconfig:"default=function-controller-docker-registry.kyma-system.svc.cluster.local:5000"`
	ExternalAddress string `envconfig:"default=registry.kyma.local"`
}
