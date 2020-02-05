/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	_ "knative.dev/pkg/client/injection/kube/client/fake"
	_ "knative.dev/pkg/injection/clients/dynamicclient/fake"

	// Link fake informers and clients accessed by our controller
	_ "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client/fake"
	_ "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/informers/sources/v1alpha1/httpsource/fake"
	_ "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/istio/client/fake"
	_ "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/istio/informers/authentication/v1alpha1/policy/fake"
	_ "knative.dev/eventing/pkg/client/injection/informers/messaging/v1alpha1/channel/fake"
	_ "knative.dev/serving/pkg/client/injection/client/fake"
	_ "knative.dev/serving/pkg/client/injection/informers/serving/v1alpha1/service/fake"
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

	// expected informers: HTTPSource, Channel, Knative Service, Policy
	ctrler := NewController(ctx, cmw)

	r := ctrler.Reconciler.(*Reconciler)
	if expect, got := 4, len(informers); got != expect {
		t.Errorf("Expected %d injected informers, got %d", expect, got)
	}
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
