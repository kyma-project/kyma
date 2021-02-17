// Code generated by injection-gen. DO NOT EDIT.

package client

import (
	"context"

	internalclientset "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset"
	rest "k8s.io/client-go/rest"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Default.RegisterClient(withClient)
}

// Key is used as the key for associating information with a context.Context.
type Key struct{}

func withClient(ctx context.Context, cfg *rest.Config) context.Context {
	return context.WithValue(ctx, Key{}, internalclientset.NewForConfigOrDie(cfg))
}

// Get extracts the internalclientset.Interface client from the context.
func Get(ctx context.Context) internalclientset.Interface {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset.Interface from context.")
	}
	return untyped.(internalclientset.Interface)
}
