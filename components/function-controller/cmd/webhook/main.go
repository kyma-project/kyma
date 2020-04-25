package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/system"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/certificates"
	"knative.dev/pkg/webhook/resourcesemantics"
	"knative.dev/pkg/webhook/resourcesemantics/defaulting"
	"knative.dev/pkg/webhook/resourcesemantics/validation"

	serverlessv1alhpa1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var types = map[schema.GroupVersionKind]resourcesemantics.GenericCRD{
	serverlessv1alhpa1.GroupVersion.WithKind("Function"): &serverlessv1alhpa1.Function{},
}

func main() {
	serviceName := os.Getenv("WEBHOOK_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "webhook-aerfio"
	}

	// Scope informers to the webhook's namespace instead of cluster-wide
	ctx := injection.WithNamespaceScope(signals.NewContext(), system.Namespace())

	// Set up a signal context with our webhook options
	ctx = webhook.WithOptions(ctx, webhook.Options{
		ServiceName: serviceName,
		Port:        8443,
		SecretName:  "webhook-certs",
	})

	config, err := getRestConfig("")
	if err != nil {
		panic(errors.New("while initializing rest config"))
	}

	sharedmain.WebhookMainWithConfig(ctx, "webhook",
		config,
		certificates.NewController,
		NewDefaultingAdmissionController,
		NewValidationAdmissionController,
	)
}

func getRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	// If we have an explicit indication of where the kubernetes config lives, read that.
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	// If not, try the in-cluster config.
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory.
	if usr, err := user.Current(); err == nil {
		if c, err := clientcmd.BuildConfigFromFlags("", filepath.Join(usr.HomeDir, ".kube", "config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not create a valid kubeconfig")
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

func NewDefaultingAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
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

func NewValidationAdmissionController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
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
