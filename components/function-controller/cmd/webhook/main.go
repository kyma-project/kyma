package main

import (
	"context"

	"github.com/pkg/errors"

	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"
	ctrl "sigs.k8s.io/controller-runtime"

	serverlessv1alhpa1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var types = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	serverlessv1alhpa1.GroupVersion.WithKind("Function"):      &serverlessv1alhpa1.Function{},
	serverlessv1alhpa1.GroupVersion.WithKind("GitRepository"): &serverlessv1alhpa1.GitRepository{},
}

type config struct {
	SystemNamespace    string `envconfig:"default=kyma-system"`
	WebhookServiceName string `envconfig:"default=serverless-webhook"`
	WebhookSecretName  string `envconfig:"default=serverless-webhook"`
	WebhookPort        int    `envconfig:"default=8443"`
}

func main() {
	cfg := &config{}
	if err := envconfig.Init(cfg); err != nil {
		panic(errors.Wrap(err, "while reading env variables"))
	}

	defaultingCfg := readDefaultingConfig()
	validationCfg := readValidationConfig()

	// Scope informers to the webhook's namespace instead of cluster-wide
	ctx := injection.WithNamespaceScope(signals.NewContext(), cfg.SystemNamespace)

	// Set up a signal context with our webhook options
	ctx = webhook.WithOptions(ctx, webhook.Options{
		ServiceName: cfg.WebhookServiceName,
		Port:        cfg.WebhookPort,
		SecretName:  cfg.WebhookSecretName,
	})

	restConfig := ctrl.GetConfigOrDie()

	sharedmain.WebhookMainWithConfig(ctx, "serverless-webhook",
		restConfig,
		certificates.NewController,
		NewDefaultingAdmissionController(defaultingCfg),
		NewValidationAdmissionController(validationCfg),
	)
}

func NewDefaultingAdmissionController(cfg *serverlessv1alhpa1.DefaultingConfig) func(ctx context.Context, _ configmap.Watcher) *controller.Impl {
	return func(ctx context.Context, _ configmap.Watcher) *controller.Impl {
		return defaulting.NewAdmissionController(ctx,

			// Name of the resource webhook.
			"defaulting.webhook.serverless.kyma-project.io",

			// The path on which to serve the webhook.
			"/defaulting",

			// The resources to validate and default.
			types,

			// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
			func(ctx context.Context) context.Context {
				return context.WithValue(ctx, serverlessv1alhpa1.DefaultingConfigKey, *cfg)
			},

			// Whether to disallow unknown fields.
			true,
		)
	}
}

func NewValidationAdmissionController(cfg *serverlessv1alhpa1.ValidationConfig) func(ctx context.Context, _ configmap.Watcher) *controller.Impl {
	return func(ctx context.Context, _ configmap.Watcher) *controller.Impl {
		return validation.NewAdmissionController(ctx,

			// Name of the resource webhook.
			"validation.webhook.serverless.kyma-project.io",

			// The path on which to serve the webhook.
			"/resource-validation",

			// The resources to validate and default.
			types,

			// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
			func(ctx context.Context) context.Context {
				return context.WithValue(ctx, serverlessv1alhpa1.ValidationConfigKey, *cfg)
			},

			// Whether to disallow unknown fields.
			true,
		)
	}
}

func readDefaultingConfig() *serverlessv1alhpa1.DefaultingConfig {
	defaultingCfg := &serverlessv1alhpa1.DefaultingConfig{}
	if err := envconfig.InitWithPrefix(defaultingCfg, "WEBHOOK_DEFAULTING"); err != nil {
		panic(errors.Wrap(err, "while reading env defaulting variables"))
	}

	functionReplicasPresets, err := serverlessv1alhpa1.ParseReplicasPresets(defaultingCfg.Function.Replicas.PresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing function resources presets"))
	}
	defaultingCfg.Function.Replicas.Presets = functionReplicasPresets

	functionResourcesPresets, err := serverlessv1alhpa1.ParseResourcePresets(defaultingCfg.Function.Resources.PresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing function resources presets"))
	}
	defaultingCfg.Function.Resources.Presets = functionResourcesPresets

	buildResourcesPresets, err := serverlessv1alhpa1.ParseResourcePresets(defaultingCfg.BuildJob.Resources.PresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing build resources presets"))
	}
	defaultingCfg.BuildJob.Resources.Presets = buildResourcesPresets

	return defaultingCfg
}

func readValidationConfig() *serverlessv1alhpa1.ValidationConfig {
	validationCfg := &serverlessv1alhpa1.ValidationConfig{}
	if err := envconfig.InitWithPrefix(validationCfg, "WEBHOOK_VALIDATION"); err != nil {
		panic(errors.Wrap(err, "while reading env defaulting variables"))
	}
	return validationCfg
}
