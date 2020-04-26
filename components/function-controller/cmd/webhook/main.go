package main

import (
	"context"
	"errors"

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
	serverlessv1alhpa1.GroupVersion.WithKind("Function"): &serverlessv1alhpa1.Function{},
}

type config struct {
	CurrentNamespace   string `envconfig:"default=kyma-system"`
	WebhookServiceName string `envconfig:"default=serverless-webhook-svc"`
	SecretName         string `envconfig:"default=serverless-webhook"`
	Port               int    `envconfig:"default=8433"`
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete

func main() {
	cfg := &config{}
	if err := envconfig.InitWithPrefix(cfg, "APP_WEBHOOK"); err != nil {
		panic(errors.New("while reading env variables"))
	}

	// Scope informers to the webhook's namespace instead of cluster-wide
	ctx := injection.WithNamespaceScope(signals.NewContext(), cfg.CurrentNamespace)

	// Set up a signal context with our webhook options
	ctx = webhook.WithOptions(ctx, webhook.Options{
		ServiceName: cfg.WebhookServiceName,
		Port:        cfg.Port,
		SecretName:  cfg.SecretName,
	})

	restConfig := ctrl.GetConfigOrDie()

	sharedmain.WebhookMainWithConfig(ctx, "serverless-webhook",
		restConfig,
		certificates.NewController,
		NewDefaultingAdmissionController,
		NewValidationAdmissionController,
	)
}

func NewDefaultingAdmissionController(ctx context.Context, _ configmap.Watcher) *controller.Impl {
	return defaulting.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"defaulting.webhook.serverless.kyma-project.io",

		// The path on which to serve the webhook.
		"/defaulting",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}

func NewValidationAdmissionController(ctx context.Context, _ configmap.Watcher) *controller.Impl {
	return validation.NewAdmissionController(ctx,

		// Name of the resource webhook.
		"validation.webhook.serverless.kyma-project.io",

		// The path on which to serve the webhook.
		"/resource-validation",

		// The resources to validate and default.
		types,

		// A function that infuses the context passed to Validate/SetDefaults with custom metadata.
		func(ctx context.Context) context.Context {
			return ctx
		},

		// Whether to disallow unknown fields.
		true,
	)
}
