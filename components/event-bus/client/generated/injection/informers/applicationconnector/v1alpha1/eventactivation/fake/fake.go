// Code generated by injection-gen. DO NOT EDIT.

package fake

import (
	"context"

	eventactivation "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/informers/applicationconnector/v1alpha1/eventactivation"
	fake "github.com/kyma-project/kyma/components/event-bus/client/generated/injection/informers/factory/fake"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
)

var Get = eventactivation.Get

func init() {
	injection.Fake.RegisterInformer(withInformer)
}

func withInformer(ctx context.Context) (context.Context, controller.Informer) {
	f := fake.Get(ctx)
	inf := f.Applicationconnector().V1alpha1().EventActivations()
	return context.WithValue(ctx, eventactivation.Key{}, inf), inf.Informer()
}
