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

	fakeeventingclientset "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	rt "knative.dev/pkg/reconciler/testing"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	applicationconnectorlistersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/lister/applicationconnector/v1alpha1"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	eventinglistersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/lister/eventing/v1alpha1"

	fakeeventbusclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/fake"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakeeventbusclientset.AddToScheme,
	fakeeventingclientset.AddToScheme,
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

func (l *Listers) GetEventBusObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeeventbusclientset.AddToScheme)
}

func (l *Listers) GetEventingObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeeventingclientset.AddToScheme)
}

func (l *Listers) GetEventActivationLister() applicationconnectorlistersv1alpha1.EventActivationLister {
	return applicationconnectorlistersv1alpha1.NewEventActivationLister(l.IndexerFor(&applicationconnectorv1alpha1.EventActivation{}))
}

func (l *Listers) GetSubscriptionLister() eventinglistersv1alpha1.SubscriptionLister {
	return eventinglistersv1alpha1.NewSubscriptionLister(l.IndexerFor(&eventingv1alpha1.Subscription{}))
}
