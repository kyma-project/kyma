package compassconnection

import (
	"context"
	"errors"
	"os"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/fake"
	clientset "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"
)

func NewFakeClient(objects ...runtime.Object) Client {
	return &fakeClientWrapper{
		fakeClient: fake.NewSimpleClientset(objects...).CompassV1alpha1().CompassConnections(),
	}
}

type fakeClientWrapper struct {
	fakeClient clientset.CompassConnectionInterface
}

func (f *fakeClientWrapper) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	compassConnection, ok := obj.(*v1alpha1.CompassConnection)
	if !ok {
		return errors.New("object is not Compass Connection")
	}

	cc, err := f.fakeClient.Get(key.Name, v1.GetOptions{})
	if err != nil {
		return err
	}

	cc.DeepCopyInto(compassConnection)
	return nil
}

func (f *fakeClientWrapper) Update(ctx context.Context, obj runtime.Object) error {
	panic("implement me")
}

func (f *fakeClientWrapper) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error {
	panic("implement me")
}

func TestMain(m *testing.M) {
	code := m.Run()

	os.Exit(code)
}
