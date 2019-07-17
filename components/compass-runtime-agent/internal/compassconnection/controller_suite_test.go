package compassconnection

import (
	"log"
	"os"
	"sync"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass"
	"github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	//"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	//clientset "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"

	//"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/fake"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//
//func NewFakeClient(objects ...runtime.Object) Client {
//	return &fakeClientWrapper{
//		fakeClient: fake.NewSimpleClientset(objects...).CompassV1alpha1().CompassConnections(),
//	}
//}
//
//type fakeClientWrapper struct {
//	fakeClient clientset.CompassConnectionInterface
//}
//
//func (f *fakeClientWrapper) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
//	compassConnection, ok := obj.(*v1alpha1.CompassConnection)
//	if !ok {
//		return errors.New("Object is not Compass Connection")
//	}
//
//	cc, err := f.fakeClient.Get(key.Name, v1.GetOptions{})
//	if err != nil {
//		return err
//	}
//
//	cc.DeepCopyInto(compassConnection)
//	return nil
//}
//
//func (f *fakeClientWrapper) Update(ctx context.Context, obj runtime.Object) error {
//	panic("implement me")
//}
//
//func (f *fakeClientWrapper) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error {
//	panic("implement me")
//}

var cfg *rest.Config

func TestMain(m *testing.M) {
	//code := m.Run()
	//
	//os.Exit(code)

	t := &envtest.Environment{
		//CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "config", "crds")},
		CRDs: []*apiextensionsv1beta1.CustomResourceDefinition{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "compassconnections.compass.kyma-project.io",
				},
				Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
					Group: "compass.kyma-project.io",
					Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
						Plural:   "compassconnections",
						Singular: "compassconnection",
						Kind:     "CompassConnection",
					},
					Version: "v1alpha1",
				},
			},
		},
	}
	apis.AddToScheme(scheme.Scheme)

	var err error
	if cfg, err = t.Start(); err != nil {
		log.Fatal(err)
	}

	code := m.Run()
	t.Stop()
	os.Exit(code)
}

// SetupTestReconcile returns a reconcile.Reconcile implementation that delegates to inner and
// writes the request to requests after Reconcile is finished.
func SetupTestReconcile(inner reconcile.Reconciler) (reconcile.Reconciler, chan reconcile.Request) {
	requests := make(chan reconcile.Request)
	fn := reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) {
		result, err := inner.Reconcile(req)
		requests <- req
		return result, err
	})
	return fn, requests
}

// StartTestManager adds recFn
func StartTestManager(mgr manager.Manager, g *gomega.GomegaWithT) (chan struct{}, *sync.WaitGroup) {
	stop := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.Expect(mgr.Start(stop)).NotTo(gomega.HaveOccurred())
	}()
	return stop, wg
}
