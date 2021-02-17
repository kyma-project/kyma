package httpsource

import (
	"context"
	"reflect"
	"testing"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	rt "knative.dev/pkg/reconciler/testing"

	. "github.com/kyma-project/kyma/components/event-sources/reconciler/testing"

	// Link fake clients accessed by reconciler.Base
	_ "knative.dev/eventing/pkg/client/injection/client/fake"
	_ "knative.dev/eventing/pkg/legacyclient/injection/client/fake"
	_ "knative.dev/pkg/client/injection/kube/client/fake"
	_ "knative.dev/pkg/injection/clients/dynamicclient/fake"

	// Link fake informers and clients accessed by our controller
	_ "knative.dev/eventing/pkg/client/injection/informers/messaging/v1alpha1/channel/fake"
	_ "knative.dev/pkg/client/injection/ducks/duck/v1/addressable/fake"
	_ "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment/fake"
	_ "knative.dev/pkg/client/injection/kube/informers/core/v1/service/fake"

	_ "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client/fake"
	_ "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/informers/sources/v1alpha1/httpsource/fake"
	_ "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/istio/client/fake"
	_ "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/istio/informers/security/v1beta1/peerauthentication/fake"
)

const adapterImageEnvVar = "HTTP_ADAPTER_IMAGE"

func TestNewController(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	defer SetEnvVar(t, adapterImageEnvVar, "some-image")()
	defer SetEnvVar(t, metrics.DomainEnv, "testing")()

	cmw := configmap.NewStaticWatcher(
		NewConfigMap("", metrics.ConfigMapName()),
		NewConfigMap("", logging.ConfigMapName()),
	)
	ctx, informers := rt.SetupFakeContext(t)

	// expected informers: HTTPSource, Channel, Deployment, PeerAuthentication, Service
	if expect, got := 5, len(informers); got != expect {
		t.Errorf("Expected %d injected informers, got %d", expect, got)
	}

	ctrler := NewController(ctx, cmw)

	r := ctrler.Reconciler.(*Reconciler)
	ensureNoNilField(reflect.ValueOf(r).Elem(), t)
}

func TestMandatoryEnvVars(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected recover to yield and error")
		}
	}()

	cmw := configmap.NewStaticWatcher()
	ctx := context.TODO()
	_ = NewController(ctx, cmw)

	t.Error("Expected function to panic")
}

// ensureNoNilField fails the test if the provided struct contains nil pointers
// or interfaces.
func ensureNoNilField(structVal reflect.Value, t *testing.T) {
	t.Helper()

	for i := 0; i < structVal.NumField(); i++ {
		f := structVal.Field(i)
		switch f.Kind() {
		case reflect.Interface, reflect.Ptr:
			if f.IsNil() {
				t.Errorf("Reconciler field %q is nil", structVal.Type().Field(i).Name)
			}
		}
	}
}
