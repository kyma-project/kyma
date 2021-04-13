package testing

import (
	"github.com/pkg/errors"
	fakeistioclientset "istio.io/client-go/pkg/clientset/versioned/fake"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/reconciler/testing"

	eaFake "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakekubeclientset.AddToScheme,
	fakeistioclientset.AddToScheme,
	eaFake.AddToScheme,
}

type Listers struct {
	sorter testing.ObjectSorter
}

func NewScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		if err := addTo(scheme); err != nil {
			panic(errors.Wrapf(err, "error while adding to scheme"))
		}
	}
	return scheme
}

func NewListers(objs []runtime.Object) Listers {
	scheme := NewScheme()

	ls := Listers{
		sorter: testing.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

func (l *Listers) indexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

func (l *Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakekubeclientset.AddToScheme)
}

func (l *Listers) GetIstioObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakeistioclientset.AddToScheme)
}

func (l *Listers) GetEAObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(eaFake.AddToScheme)
}
