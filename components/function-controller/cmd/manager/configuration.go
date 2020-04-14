package main

import (
	"os"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers"
	"github.com/kyma-project/kyma/components/function-controller/pkg/envars"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type dockerCgf struct {
	dockerRegistryPort            int
	dockerRegistryFqdn            string
	dockerRegistryExternalAddress string
	dockerRegistryName            string
}

type configuration struct {
	metricsAddr             string
	logDevModeEnabled       bool
	leaderElectionEnabled   bool
	leaderElectionNamespace string
	leaderElectionID        string
	maxConcurrentReconciles int
	tektonRequestsCPU       resource.Quantity
	tektonRequestsMem       resource.Quantity
	tektonLimitsCPU         resource.Quantity
	tektonLimitsMem         resource.Quantity
	runtimeConfigmapName    string
	imagePullSecretName     string
	imagePullAccount        string
	requeueDuration         time.Duration
	dockerCgf
}

func (appCfg *configuration) fnReconcilerCfg() *controllers.FnReconcilerCfg {
	return &controllers.FnReconcilerCfg{
		Limits: &corev1.ResourceList{
			corev1.ResourceCPU:    appCfg.tektonLimitsCPU,
			corev1.ResourceMemory: appCfg.tektonLimitsMem,
		},
		Requests: &corev1.ResourceList{
			corev1.ResourceCPU:    appCfg.tektonRequestsCPU,
			corev1.ResourceMemory: appCfg.tektonRequestsMem,
		},
		MaxConcurrentReconciles: appCfg.maxConcurrentReconciles,
		DockerCfg: controllers.DockerCfg{
			DockerRegistryPort:            appCfg.dockerRegistryPort,
			DockerRegistryFqdn:            appCfg.dockerRegistryFqdn,
			DockerRegistryExternalAddress: appCfg.dockerRegistryExternalAddress,
			DockerRegistryName:            appCfg.dockerRegistryName,
		},
		ImagePullSecretName: appCfg.imagePullSecretName,
		ImagePullAccount:    appCfg.imagePullAccount,
		RuntimeConfigmap:    appCfg.runtimeConfigmapName,
	}
}

type option = func() error

func loadConfigurationOrDie(appCfg *configuration) {
	err := loadConfiguration(appCfg)
	if err != nil {
		setupLog.Error(err, "Unable to read application configuration")
		os.Exit(1)
	}
}

func loadConfiguration(cfg *configuration) error {
	options := []option{
		envars.BuildOptionParseEvnStrWithDefault(
			&cfg.metricsAddr,
			"METRICS_ADDR",
			&defaultMetricsAddr),

		envars.BuildOptionParseEvnBool(
			&cfg.logDevModeEnabled,
			"LOG_DEV_MODE_ENABLED"),

		envars.BuildOptionParseEvnBool(
			&cfg.leaderElectionEnabled,
			"LEADER_ELECTION_ENABLED"),

		envars.BuildOptionParseEvnStrWithDefault(
			&cfg.leaderElectionNamespace,
			"LEADER_ELECTION_NAMESPACE",
			&empty),

		envars.BuildOptionParseEvnStrWithDefault(
			&cfg.leaderElectionID,
			"LEADER_ELECTION_ID",
			&empty),

		envars.BuildOptionParseEvnQuantityWithDefault(
			&cfg.tektonRequestsCPU,
			"TEKTON_REQUESTS_CPU",
			&controllers.DefaultTektonRequestsCPU),

		envars.BuildOptionParseEvnQuantityWithDefault(
			&cfg.tektonRequestsMem,
			"TEKTON_REQUESTS_MEMORY",
			&controllers.DefaultTektonRequestsMem),

		envars.BuildOptionParseEvnQuantityWithDefault(
			&cfg.tektonLimitsCPU,
			"TEKTON_LIMITS_CPU",
			&controllers.DefaultTektonLimitsCPU),

		envars.BuildOptionParseEvnQuantityWithDefault(
			&cfg.tektonLimitsMem,
			"TEKTON_LIMITS_MEMORY",
			&controllers.DefaultTektonLimitsMem),

		envars.BuildOptionParseEvnIntWithDefault(
			&cfg.dockerRegistryPort,
			"DOCKER_REGISTRY_PORT",
			&controllers.DefaultDockerRegistryPort),

		envars.BuildOptionParseEvnStrWithDefault(
			&cfg.dockerRegistryFqdn,
			"DOCKER_REGISTRY_FQDN",
			&controllers.DefaultDockerRegistryFqdn),

		envars.BuildOptionParseEvnStrWithDefault(
			&cfg.dockerRegistryName,
			"DOCKER_REGISTRY_NAME",
			&empty),

		envars.BuildOptionParseEvnStrWithDefault(
			&cfg.dockerRegistryExternalAddress,
			"DOCKER_REGISTRY_EXTERNAL_ADDRESS",
			&controllers.DefaultDockerRegistryExternalAddress),

		envars.BuildOptionParseEvnStrWithDefault(
			&cfg.runtimeConfigmapName,
			"RUNTIME_CONFIGMAP",
			&controllers.DefaultRuntimeConfigmapName),

		envars.BuildOptionParseEvnStrWithDefault(
			&cfg.imagePullAccount,
			"IMAGE_PULL_ACCOUNT",
			&controllers.DefaultImagePullAccount),

		envars.BuildOptionParseEvnDurationWithDefault(
			&cfg.requeueDuration,
			"REQUEUE_DURATION",
			&controllers.DefaultRequeueDuration),
	}
	for _, option := range options {
		if err := option(); err != nil {
			return err
		}
	}
	return nil
}
