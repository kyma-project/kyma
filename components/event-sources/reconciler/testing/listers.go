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

package testing

import (
	pkgerrors "github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"

	authv1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
	fakeistioclientset "istio.io/client-go/pkg/clientset/versioned/fake"
	authenticationlistersv1alpha1 "istio.io/client-go/pkg/listers/authentication/v1alpha1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	fakeeventingclientset "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	messaginglistersv1alpha1 "knative.dev/eventing/pkg/client/listers/messaging/v1alpha1"
	rt "knative.dev/pkg/reconciler/testing"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	fakeservingclientset "knative.dev/serving/pkg/client/clientset/versioned/fake"
	servinglistersv1alpha1 "knative.dev/serving/pkg/client/listers/serving/v1alpha1"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	fakesourcesclientset "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/fake"
	sourceslistersv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/lister/sources/v1alpha1"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakeservingclientset.AddToScheme,
	fakesourcesclientset.AddToScheme,
	fakeeventingclientset.AddToScheme,
	fakeistioclientset.AddToScheme,
}

type Listers struct {
	sorter rt.ObjectSorter
}

func NewListers(scheme *runtime.Scheme, objs []runtime.Object) Listers {
	ls := Listers{
		sorter: rt.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	sb := runtime.NewSchemeBuilder(clientSetSchemes...)
	if err := sb.AddToScheme(scheme); err != nil {
		panic(pkgerrors.Wrap(err, "building Scheme"))
	}

	return scheme
}

// IndexerFor returns the indexer for the given object.
func (l *Listers) IndexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

func (l *Listers) GetSourcesObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakesourcesclientset.AddToScheme)
}

func (l *Listers) GetServingObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeservingclientset.AddToScheme)
}

func (l *Listers) GetEventingObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeeventingclientset.AddToScheme)
}

func (l *Listers) GetIstioObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeistioclientset.AddToScheme)
}

func (l *Listers) GetHTTPSourceLister() sourceslistersv1alpha1.HTTPSourceLister {
	return sourceslistersv1alpha1.NewHTTPSourceLister(l.IndexerFor(&sourcesv1alpha1.HTTPSource{}))
}

func (l *Listers) GetServiceLister() servinglistersv1alpha1.ServiceLister {
	return servinglistersv1alpha1.NewServiceLister(l.IndexerFor(&servingv1alpha1.Service{}))
}

func (l *Listers) GetChannelLister() messaginglistersv1alpha1.ChannelLister {
	return messaginglistersv1alpha1.NewChannelLister(l.IndexerFor(&messagingv1alpha1.Channel{}))
}

func (l *Listers) GetPolicyLister() authenticationlistersv1alpha1.PolicyLister {
	return authenticationlistersv1alpha1.NewPolicyLister(l.IndexerFor(&authv1alpha1.Policy{}))
}
