package testing

import (
	pkgerrors "github.com/pkg/errors"
	fakeistioclientset "istio.io/client-go/pkg/clientset/versioned/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"
	appslistersv1 "k8s.io/client-go/listers/apps/v1"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	fakeeventingclientset "knative.dev/eventing/pkg/client/clientset/versioned/fake"
	messaginglistersv1alpha1 "knative.dev/eventing/pkg/client/listers/messaging/v1alpha1"
	fakelegacyclientset "knative.dev/eventing/pkg/legacyclient/clientset/versioned/fake"
	_ "knative.dev/pkg/client/injection/ducks/duck/v1/addressable/fake"
	rt "knative.dev/pkg/reconciler/testing"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	securitylistersv1alpha1 "istio.io/client-go/pkg/listers/security/v1beta1"

	sourcesv1alpha1 "github.com/kyma-project/kyma/components/event-sources/apis/sources/v1alpha1"
	fakesourcesclientset "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/fake"
	sourceslistersv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/lister/sources/v1alpha1"
)

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakesourcesclientset.AddToScheme,
	fakeeventingclientset.AddToScheme,
	fakelegacyclientset.AddToScheme,
	fakeistioclientset.AddToScheme,
	fakekubeclientset.AddToScheme,
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

func (l *Listers) GetCoreObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(corev1.AddToScheme)
}

func (l *Listers) GetAppsObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(appsv1.AddToScheme)
}

func (l *Listers) GetLegacyObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakelegacyclientset.AddToScheme)
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

func (l *Listers) GetDeploymentLister() appslistersv1.DeploymentLister {
	return appslistersv1.NewDeploymentLister(l.IndexerFor(&appsv1.Deployment{}))
}

func (l *Listers) GetChannelLister() messaginglistersv1alpha1.ChannelLister {
	return messaginglistersv1alpha1.NewChannelLister(l.IndexerFor(&messagingv1alpha1.Channel{}))
}

func (l *Listers) GetPeerAuthenticationLister() securitylistersv1alpha1.PeerAuthenticationLister {
	return securitylistersv1alpha1.NewPeerAuthenticationLister(l.IndexerFor(&securityv1beta1.PeerAuthentication{}))
}

func (l *Listers) GetServiceLister() corelistersv1.ServiceLister {
	return corelistersv1.NewServiceLister(l.IndexerFor(&corev1.Service{}))
}
