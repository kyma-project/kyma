package webhook

import (
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	SystemNamespace string `envconfig:"default=kyma-system"`
	ServiceName     string `envconfig:"default=serverless-webhook"`
	SecretName      string `envconfig:"default=serverless-webhook"`
	Port            int    `envconfig:"default=8443"`
	LogConfigPath   string `envconfig:"default=/appdata/log_config.yaml"`
	ConfigPath      string `envconfig:"default=/appdata/config.yaml"`
}

func ReadDefaultingConfigV1Alpha2OrDie() *serverlessv1alpha2.DefaultingConfig {
	defaultingCfg := &serverlessv1alpha2.DefaultingConfig{}
	if err := envconfig.InitWithPrefix(defaultingCfg, "WEBHOOK_DEFAULTING"); err != nil {
		panic(errors.Wrap(err, "while reading env defaulting variables"))
	}

	functionReplicasPresets, err := serverlessv1alpha2.ParseReplicasPresets(defaultingCfg.Function.Replicas.PresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing function replicas presets"))
	}
	defaultingCfg.Function.Replicas.Presets = functionReplicasPresets

	functionResourcesPresets, err := serverlessv1alpha2.ParseResourcePresets(defaultingCfg.Function.Resources.PresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing function resources presets"))
	}
	defaultingCfg.Function.Resources.Presets = functionResourcesPresets

	buildResourcesPresets, err := serverlessv1alpha2.ParseResourcePresets(defaultingCfg.BuildJob.Resources.PresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing build resources presets"))
	}
	defaultingCfg.BuildJob.Resources.Presets = buildResourcesPresets

	runtimePresets, err := serverlessv1alpha2.ParseRuntimePresets(defaultingCfg.Function.Resources.RuntimePresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing runtime preset"))
	}
	defaultingCfg.Function.Resources.RuntimePresets = runtimePresets

	return defaultingCfg
}

func ReadValidationConfigV1Alpha2OrDie() *serverlessv1alpha2.ValidationConfig {
	validationCfg := &serverlessv1alpha2.ValidationConfig{}
	if err := envconfig.InitWithPrefix(validationCfg, "WEBHOOK_VALIDATION"); err != nil {
		panic(errors.Wrap(err, "while reading env defaulting variables"))
	}
	return validationCfg
}
