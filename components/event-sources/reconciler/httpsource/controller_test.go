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
	"os"
	"reflect"
	"testing"

	"knative.dev/pkg/configmap"
	rt "knative.dev/pkg/reconciler/testing"

	// Link fake clients accessed by reconciler.Base
	_ "knative.dev/eventing/pkg/client/injection/client/fake"
	_ "knative.dev/pkg/client/injection/kube/client/fake"
	_ "knative.dev/pkg/injection/clients/dynamicclient/fake"

	// Link fake informers and clients accessed by our controller
	_ "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/client/fake"
	_ "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/informers/sources/v1alpha1/httpsource/fake"
	_ "knative.dev/serving/pkg/client/injection/client/fake"
	_ "knative.dev/serving/pkg/client/injection/informers/serving/v1/service/fake"
)

func TestNewController(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	defer setAdapterImageEnvVar("some-image", t)()

	cmw := configmap.NewStaticWatcher()
	ctx, informers := rt.SetupFakeContext(t)

	// expected infomers: HTTPSource and Knative Service
	if expect, got := 2, len(informers); got != expect {
		t.Errorf("Expected %d injected informers, got %d", expect, got)
	}

	ctrler := NewController(ctx, cmw)
	r := ctrler.Reconciler.(*Reconciler)

	ensureNoNilField(reflect.ValueOf(r).Elem(), t)
}

func TestGetAdapterImage(t *testing.T) {
	t.Run("Returns value of env var if set", func(t *testing.T) {
		expectVal := "test"

		defer setAdapterImageEnvVar(expectVal, t)()

		ai := getAdapterImage()
		if ai != expectVal {
			t.Errorf("Expected value of env var %s to be %q, got %q", adapterImageEnvVar, expectVal, ai)
		}
	})

	t.Run("Panics if env var unset", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected recover to yield and error")
			}
		}()
		_ = getAdapterImage()
		t.Error("Expected function to panic")
	})
}

// setAdapterImageEnvVar sets the adapter image env var and returns a function
// that can be deferred to unset that variable.
func setAdapterImageEnvVar(val string, t *testing.T) (unset func() error) {
	t.Helper()

	if err := os.Setenv(adapterImageEnvVar, val); err != nil {
		t.Errorf("Failed to set env var %s: %v", adapterImageEnvVar, err)
	}

	return func() error {
		return os.Unsetenv(adapterImageEnvVar)
	}
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
