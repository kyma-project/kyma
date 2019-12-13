// Code generated by injection-gen. DO NOT EDIT.

package factory

import (
	"context"

	externalversions "github.com/kyma-project/kyma/components/event-bus/client/generated/informer/externalversions"
	client "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/client"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Default.RegisterInformerFactory(withInformerFactory)
}

// Key is used as the key for associating information with a context.Context.
type Key struct{}

func withInformerFactory(ctx context.Context) context.Context {
	c := client.Get(ctx)
	opts := make([]externalversions.SharedInformerOption, 0, 1)
	if injection.HasNamespaceScope(ctx) {
		opts = append(opts, externalversions.WithNamespace(injection.GetNamespaceScope(ctx)))
	}
	return context.WithValue(ctx, Key{},
		externalversions.NewSharedInformerFactoryWithOptions(c, controller.GetResyncPeriod(ctx), opts...))
}

// Get extracts the InformerFactory from the context.
func Get(ctx context.Context) externalversions.SharedInformerFactory {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch github.com/kyma-project/kyma/components/event-bus/client/generated/informer/externalversions.SharedInformerFactory from context.")
	}
	return untyped.(externalversions.SharedInformerFactory)
}
