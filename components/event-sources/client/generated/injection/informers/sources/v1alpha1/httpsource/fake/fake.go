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

// Code generated by injection-gen. DO NOT EDIT.

package fake

import (
	"context"

	fake "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/informers/factory/fake"
	httpsource "github.com/kyma-project/kyma/components/event-sources/client/generated/injection/informers/sources/v1alpha1/httpsource"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
)

var Get = httpsource.Get

func init() {
	injection.Fake.RegisterInformer(withInformer)
}

func withInformer(ctx context.Context) (context.Context, controller.Informer) {
	f := fake.Get(ctx)
	inf := f.Sources().V1alpha1().HTTPSources()
	return context.WithValue(ctx, httpsource.Key{}, inf), inf.Informer()
}
